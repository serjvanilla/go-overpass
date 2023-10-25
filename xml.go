package overpass

import (
	"encoding/xml"
	"fmt"
	"time"
)

type ActionType string

const (
	ActionTypeCreate ActionType = "create"
	ActionTypeModify ActionType = "modify"
	ActionTypeDelete ActionType = "delete"
)

type Attributes struct {
	ID        int64      `xml:"id,attr"`
	Timestamp *time.Time `xml:"timestamp,attr"`
	Version   int64      `xml:"version,attr"`
	Changeset int64      `xml:"changeset,attr"`
	User      string     `xml:"user,attr"`
	UID       int64      `xml:"uid,attr"`
	Visible   bool       `xml:"visible,attr"`
	Tags      []Tag      `xml:"tag"`
}

type Tag struct {
	K string `xml:"k,attr"`
	V string `xml:"v,attr"`
}

type NodeXml struct {
	Attributes
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
}

type WayXml struct {
	Attributes
	Bounds *Bounds `xml:"bounds"`
	Nodes  []*struct {
		Ref int64   `xml:"ref,attr"`
		Lat float64 `xml:"lat,attr"`
		Lon float64 `xml:"lon,attr"`
	} `xml:"nd"`
}

type RelationXml struct {
	Attributes
	Bounds  *Bounds `xml:"bounds"`
	Members []struct {
		Type ElementType `xml:"type,attr"`
		Ref  int64       `xml:"ref,attr"`
		Role string      `xml:"role,attr"`
	} `xml:"member"`
}

type Bounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
}

type Element struct {
	Node     *NodeXml     `xml:"node"`
	Way      *WayXml      `xml:"way"`
	Relation *RelationXml `xml:"relation"`
}

type overpassXmlResponse struct {
	Version   string `xml:"version,attr"`
	Generator string `xml:"generator,attr"`
	Note      string `xml:"note"`
	Meta      struct {
		OsmBase time.Time  `xml:"osm_base,attr"`
		Areas   *time.Time `xml:"areas,attr"`
	} `xml:"meta"`
	Bounds    *Bounds        `xml:"bounds"`
	Nodes     []*NodeXml     `xml:"node"`
	Ways      []*WayXml      `xml:"way"`
	Relations []*RelationXml `xml:"relation"`
	Actions   []*struct {
		Type     ActionType   `xml:"type,attr"`
		Old      *Element     `xml:"old"`
		New      *Element     `xml:"new"`
		Node     *NodeXml     `xml:"node"`
		Way      *WayXml      `xml:"way"`
		Relation *RelationXml `xml:"relation"`
	} `xml:"action"`
}

func unmarshalXml(body []byte) (Result, error) {
	var overpassRes overpassXmlResponse
	if err := xml.Unmarshal(body, &overpassRes); err != nil {
		return Result{}, fmt.Errorf("overpass engine error: %w", err)
	}

	result := Result{
		Timestamp: overpassRes.Meta.OsmBase,
		Nodes:     make(map[int64]*Node),
		Ways:      make(map[int64]*Way),
		Relations: make(map[int64]*Relation),
	}

	if overpassRes.Actions == nil {
		result.Count = len(overpassRes.Nodes) + len(overpassRes.Ways) + len(overpassRes.Relations)
		for _, el := range overpassRes.Nodes {
			node := result.getNode(el.ID)
			result.processNode(el, node)
		}

		for _, el := range overpassRes.Ways {
			way := result.getWay(el.ID)
			result.processWay(el, way)
		}

		for _, el := range overpassRes.Relations {
			relation := result.getRelation(el.ID)
			result.processRelation(el, relation)
		}
	} else {
		result.Count = len(overpassRes.Actions)
		result.OldNodes = make(map[int64]*Node)
		result.OldWays = make(map[int64]*Way)
		result.OldRelations = make(map[int64]*Relation)
		for _, el := range overpassRes.Actions {
			switch el.Type {
			case ActionTypeCreate:
				if result.Create == nil {
					result.Create = &Create{}
				}
				if el.Node != nil {
					node := result.getNode(el.Node.ID)
					result.processNode(el.Node, node)
					if result.Create.Nodes == nil {
						result.Create.Nodes = make(map[int64]*Node)
					}
					result.Create.Nodes[el.Node.ID] = node
				}
				if el.Way != nil {
					way := result.getWay(el.Way.ID)
					result.processWay(el.Way, way)
					if result.Create.Ways == nil {
						result.Create.Ways = make(map[int64]*Way)
					}
					result.Create.Ways[el.Way.ID] = way
				}
				if el.Relation != nil {
					relation := result.getRelation(el.Relation.ID)
					result.processRelation(el.Relation, relation)
					if result.Create.Relations == nil {
						result.Create.Relations = make(map[int64]*Relation)
					}
					result.Create.Relations[el.Relation.ID] = relation
				}

			case ActionTypeModify:
				if result.Modify == nil {
					result.Modify = &Modify{}
				}
				if el.Old.Node != nil {
					node := result.getOldNode(el.Old.Node.ID)
					result.processNode(el.Old.Node, node)
					if result.Modify.Nodes == nil {
						result.Modify.Nodes = make(map[int64]map[string]*Node)
					}
					result.Modify.Nodes[el.Old.Node.ID] = map[string]*Node{"old": node}
				}
				if el.New.Node != nil {
					node := result.getNode(el.New.Node.ID)
					result.processNode(el.New.Node, node)
					result.Modify.Nodes[el.New.Node.ID]["new"] = node
				}
				if el.Old.Way != nil {
					way := result.getOldWay(el.Old.Way.ID)
					result.processWay(el.Old.Way, way)
					if result.Modify.Ways == nil {
						result.Modify.Ways = make(map[int64]map[string]*Way)
					}
					result.Modify.Ways[el.Old.Way.ID] = map[string]*Way{"old": way}
				}
				if el.New.Way != nil {
					way := result.getWay(el.New.Way.ID)
					result.processWay(el.New.Way, way)
					result.Modify.Ways[el.New.Way.ID]["new"] = way
				}
				if el.Old.Relation != nil {
					relation := result.getOldRelation(el.Old.Relation.ID)
					result.processRelation(el.Old.Relation, relation)
					if result.Modify.Relations == nil {
						result.Modify.Relations = make(map[int64]map[string]*Relation)
					}
					result.Modify.Relations[el.Old.Relation.ID] = map[string]*Relation{"old": relation}
				}
				if el.New.Relation != nil {
					relation := result.getRelation(el.New.Relation.ID)
					result.processRelation(el.New.Relation, relation)
					result.Modify.Relations[el.Old.Relation.ID]["new"] = relation
				}

			case ActionTypeDelete:
				if result.Delete == nil {
					result.Delete = &Delete{}
				}
				if el.Old.Node != nil {
					node := result.getOldNode(el.Old.Node.ID)
					result.processNode(el.Old.Node, node)
					if result.Delete.Nodes == nil {
						result.Delete.Nodes = make(map[int64]*Node)
					}
					result.Delete.Nodes[el.Old.Node.ID] = node
				}
				if el.New != nil && el.New.Node != nil {
					result.Delete.Nodes[el.New.Node.ID].Visible = &el.New.Node.Visible
				}
				if el.Old.Way != nil {
					way := result.getOldWay(el.Old.Way.ID)
					result.processWay(el.Old.Way, way)
					if result.Delete.Ways == nil {
						result.Delete.Ways = make(map[int64]*Way)
					}
					result.Delete.Ways[el.Old.Way.ID] = way
				}
				if el.New != nil && el.New.Way != nil {
					result.Delete.Ways[el.New.Way.ID].Visible = &el.New.Way.Visible
				}
				if el.Old.Relation != nil {
					relation := result.getOldRelation(el.Old.Relation.ID)
					result.processRelation(el.Old.Relation, relation)
					if result.Delete.Relations == nil {
						result.Delete.Relations = make(map[int64]*Relation)
					}
					result.Delete.Relations[el.Old.Relation.ID] = relation
				}
				if el.New != nil && el.New.Relation != nil {
					result.Delete.Relations[el.New.Relation.ID].Visible = &el.New.Relation.Visible
				}
			}
		}
	}

	return result, nil
}

func (r *Result) processNode(el *NodeXml, node *Node) {
	meta := Meta{
		ID:        el.ID,
		Timestamp: el.Timestamp,
		Version:   el.Version,
		Changeset: el.Changeset,
		User:      el.User,
		UID:       el.UID,
	}
	*node = Node{
		Meta: meta,
		Lat:  el.Lat,
		Lon:  el.Lon,
	}
	if len(el.Tags) > 0 {
		node.Meta.Tags = make(map[string]string)
	}
	for _, tag := range el.Tags {
		node.Meta.Tags[tag.K] = tag.V
	}
}

func (r *Result) processWay(el *WayXml, way *Way) {
	meta := Meta{
		ID:        el.ID,
		Timestamp: el.Timestamp,
		Version:   el.Version,
		Changeset: el.Changeset,
		User:      el.User,
		UID:       el.UID,
	}
	*way = Way{
		Meta:     meta,
		Nodes:    make([]*Node, len(el.Nodes)),
		Geometry: make([]Point, len(el.Nodes)),
	}
	for idx, n := range el.Nodes {
		way.Nodes[idx] = r.getNode(n.Ref)
		way.Geometry[idx].Lat = n.Lat
		way.Geometry[idx].Lon = n.Lon
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
	if len(el.Tags) > 0 {
		way.Meta.Tags = make(map[string]string)
	}
	for _, tag := range el.Tags {
		way.Meta.Tags[tag.K] = tag.V
	}
}

func (r *Result) processRelation(el *RelationXml, relation *Relation) {
	meta := Meta{
		ID:        el.ID,
		Timestamp: el.Timestamp,
		Version:   el.Version,
		Changeset: el.Changeset,
		User:      el.User,
		UID:       el.UID,
	}
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
			relationMember.Node = r.getNode(member.Ref)
		case ElementTypeWay:
			relationMember.Way = r.getWay(member.Ref)
		case ElementTypeRelation:
			relationMember.Relation = r.getRelation(member.Ref)
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
	if len(el.Tags) > 0 {
		relation.Meta.Tags = make(map[string]string)
	}
	for _, tag := range el.Tags {
		relation.Meta.Tags[tag.K] = tag.V
	}
}
