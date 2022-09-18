package overpass

import "time"

// ElementType represents possible types for Overpass response elements.
type ElementType string

// Possible values are node, way and relation.
const (
	ElementTypeNode     ElementType = "node"
	ElementTypeWay      ElementType = "way"
	ElementTypeRelation ElementType = "relation"
)

// Meta contains fields common for all OSM types.
type Meta struct {
	ID        int64
	Timestamp *time.Time
	Version   int64
	Changeset int64
	User      string
	UID       int64
	Tags      map[string]string
}

// Node represents OSM node type.
type Node struct {
	Meta
	Lat float64
	Lon float64
}

// Way represents OSM way type.
type Way struct {
	Meta
	Nodes    []*Node
	Bounds   *Box
	Geometry []Point
}

type Point struct {
	Lat, Lon float64
}

// Relation represents OSM relation type.
type Relation struct {
	Meta
	Members []RelationMember
	Bounds  *Box
}

type Box struct {
	Min, Max Point
}

// RelationMember represents OSM relation member type.
type RelationMember struct {
	Type     ElementType
	Node     *Node
	Way      *Way
	Relation *Relation
	Role     string
}

// Result returned by Query and contains parsed result of Overpass query.
type Result struct {
	Timestamp time.Time
	Count     int
	Nodes     map[int64]*Node
	Ways      map[int64]*Way
	Relations map[int64]*Relation
}
