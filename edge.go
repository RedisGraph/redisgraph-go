package redisgraph

import (
	"fmt"
	"strings"
)

// Edge represents an edge connecting two nodes in the graph.
type Edge struct {
	ID          uint64
	Relation    string
	Source      *Node
	Destination *Node
	Properties  map[string]interface{}
	srcNodeID   uint64
	destNodeID  uint64
	graph       *Graph
}

// EdgeNew create a new Edge
func EdgeNew(relation string, srcNode *Node, destNode *Node, properties map[string]interface{}) *Edge {
	p := properties
	if p == nil {
		p = make(map[string]interface{})
	}

	return &Edge{
		Relation:    relation,
		Source:      srcNode,
		Destination: destNode,
		Properties:  p,
		graph:       nil,
	}
}

// SetProperty assign a new property to edge
func (e *Edge) SetProperty(key string, value interface{}) {
	e.Properties[key] = value
}

// GetProperty retrieves property from edge
func (e *Edge) GetProperty(key string) interface{} {
	v, _ := e.Properties[key]
	return v
}

// SourceNodeID returns edge source node ID
func (e Edge) SourceNodeID() uint64 {
	if e.Source != nil {
		return e.Source.ID
	} else {
		return e.srcNodeID
	}
}

// DestNodeID returns edge destination node ID
func (e Edge) DestNodeID() uint64 {
	if e.Source != nil {
		return e.Destination.ID
	} else {
		return e.destNodeID
	}
}

// Returns a string representation of edge
func (e Edge) String() string {
	if len(e.Properties) == 0 {
		return "{}"
	}

	p := make([]string, 0, len(e.Properties))
	for k, v := range e.Properties {
		p = append(p, fmt.Sprintf("%s:%v", k, ToString(v)))
	}

	s := fmt.Sprintf("{%s}", strings.Join(p, ","))
	return s
}

// Encode makes Edge satisfy the Stringer interface
func (e Edge) Encode() string {
	s := []string{"(", e.Source.Alias, ")"}

	s = append(s, "-[")

	if e.Relation != "" {
		s = append(s, ":", e.Relation)
	}

	if len(e.Properties) > 0 {
		p := make([]string, 0, len(e.Properties))
		for k, v := range e.Properties {
			p = append(p, fmt.Sprintf("%s:%v", k, ToString(v)))
		}

		s = append(s, "{", strings.Join(p, ","), "}")
	}

	s = append(s, "]->", "(", e.Destination.Alias, ")")

	return strings.Join(s, "")
}
