package overpass

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"time"
)

var (
	// ErrHTTPError is returned on HTTP client errors
	ErrHTTPError = errors.New("HTTP error")
	// ErrOverpassError is returned when overpass returns not valid response
	ErrOverpassError = errors.New("Overpass engine error")
)

type overpassResponse struct {
	OSM3S struct {
		TimestampOSMBase time.Time `json:"timestamp_osm_base"`
	} `json:"osm3s"`
	Elements []overpassResponseElement `json:"elements"`
}

type overpassResponseElement struct {
	Type      ElementType `json:"type"`
	ID        int64       `json:"id"`
	Lat       float64     `json:"lat"`
	Lon       float64     `json:"lon"`
	Timestamp *time.Time  `json:"timestamp"`
	Version   int64       `json:"version"`
	Changeset int64       `json:"changeset"`
	User      string      `json:"user"`
	UID       int64       `json:"uid"`
	Nodes     []int64     `json:"nodes"`
	Members   []struct {
		Type ElementType `json:"type"`
		Ref  int64       `json:"ref"`
		Role string      `json:"role"`
	} `json:"members"`
	Tags map[string]string `json:"tags"`
}

// Query send request to OverpassAPI with provided querystring.
func (c *Client) Query(query string) (Result, error) {
	body, err := c.httpPost(query)
	if err != nil {
		return Result{}, err
	}

	var overpassRes overpassResponse
	err = json.Unmarshal(body, &overpassRes)
	if err != nil {
		return Result{}, ErrOverpassError
	}

	result := Result{
		Timestamp: overpassRes.OSM3S.TimestampOSMBase,
		Count:     len(overpassRes.Elements),
		Nodes:     make(map[int64]*Node),
		Ways:      make(map[int64]*Way),
		Relations: make(map[int64]*Relation),
	}

	for _, el := range overpassRes.Elements {
		meta := Meta{
			ID:        el.ID,
			Timestamp: el.Timestamp,
			Version:   el.Version,
			Changeset: el.Changeset,
			User:      el.User,
			UID:       el.UID,
			Tags:      el.Tags,
		}
		switch el.Type {
		case ElementTypeNode:
			node := result.getNode(el.ID)
			*node = Node{
				Meta: meta,
				Lat:  el.Lat,
				Lon:  el.Lon,
			}
		case ElementTypeWay:
			way := result.getWay(el.ID)
			*way = Way{
				Meta:  meta,
				Nodes: make([]*Node, len(el.Nodes)),
			}
			for idx, nodeID := range el.Nodes {
				way.Nodes[idx] = result.getNode(nodeID)
			}
		case ElementTypeRelation:
			relation := result.getRelation(el.ID)
			*relation = Relation{
				Meta:    meta,
				Members: make([]RelationMember, len(el.Members)),
			}
			for idx, member := range el.Members {
				relationMember := RelationMember{
					Type: member.Type,
					Role: member.Role,
				}
				switch member.Type {
				case ElementTypeNode:
					relationMember.Node = result.getNode(member.Ref)
				case ElementTypeWay:
					relationMember.Way = result.getWay(member.Ref)
				case ElementTypeRelation:
					relationMember.Relation = result.getRelation(member.Ref)
				}
				relation.Members[idx] = relationMember
			}
		}
	}

	return result, nil
}

func (c *Client) httpPost(query string) ([]byte, error) {
	<-c.semaphore
	defer func() { c.semaphore <- struct{}{} }()

	resp, err := c.httpClient.PostForm(
		c.apiEndpoint,
		url.Values{"data": []string{query}},
	)
	if err != nil {
		return nil, ErrHTTPError
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ErrOverpassError
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrHTTPError
	}
	return body, nil
}

// Query runs query with default client.
func Query(query string) (Result, error) {
	return DefaultClient.Query(query)
}
