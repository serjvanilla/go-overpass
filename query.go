package overpass

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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
	Geometry []struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"geometry"`
	Bounds *struct {
		MinLat float64 `json:"minlat"`
		MinLon float64 `json:"minlon"`
		MaxLat float64 `json:"maxlat"`
		MaxLon float64 `json:"maxlon"`
	} `json:"bounds"`
	Tags map[string]string `json:"tags"`
}

// Query send request to OverpassAPI with provided querystring.
func (c *Client) Query(query string) (Result, error) {
	body, err := c.httpPost(query)
	if err != nil {
		return Result{}, err
	}

	return unmarshal(body)
}

func unmarshal(body []byte) (Result, error) {
	var overpassRes overpassResponse
	if err := json.Unmarshal(body, &overpassRes); err != nil {
		return Result{}, fmt.Errorf("overpass engine error: %w", err)
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
				Meta:     meta,
				Nodes:    make([]*Node, len(el.Nodes)),
				Geometry: make([]Point, len(el.Geometry)),
			}
			for idx, nodeID := range el.Nodes {
				way.Nodes[idx] = result.getNode(nodeID)
			}
			if el.Bounds != nil {
				way.Bounds = &Box{
					Min: Point{
						Lat: el.Bounds.MinLat,
						Lon: el.Bounds.MinLon,
					},
					Max: Point{
						Lat: el.Bounds.MaxLat,
						Lon: el.Bounds.MaxLon,
					},
				}
			}
			for idx, geo := range el.Geometry {
				way.Geometry[idx].Lat = geo.Lat
				way.Geometry[idx].Lon = geo.Lon
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
			if el.Bounds != nil {
				relation.Bounds = &Box{
					Min: Point{
						Lat: el.Bounds.MinLat,
						Lon: el.Bounds.MinLon,
					},
					Max: Point{
						Lat: el.Bounds.MaxLat,
						Lon: el.Bounds.MaxLon,
					},
				}
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
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("overpass engine error: %w", &ServerError{resp.StatusCode, body})
	}

	return body, nil
}

// Query runs query with default client.
func Query(query string) (Result, error) {
	return DefaultClient.Query(query)
}

type ServerError struct {
	StatusCode int
	Body       []byte
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode))
}
