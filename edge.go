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

func (e *Edge) SetProperty(key string, value interface{}) {
	e.Properties[key] = value
}

func (e *Edge) GetProperty(key string) interface{} {
	v, _ := e.Properties[key]
	return v
}

func (e Edge) SourceNodeID() uint64 {
	if e.Source != nil {
		return e.Source.ID
	} else {
		return e.srcNodeID
	}
}

func (e Edge) DestNodeID() uint64 {
	if e.Source != nil {
		return e.Destination.ID
	} else {
		return e.destNodeID
	}
}

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

		s = append(s, "{")
		s = append(s, strings.Join(p, ","))
		s = append(s, "}")
	}

	s = append(s, "]->")
	s = append(s, "(", e.Destination.Alias, ")")

	return strings.Join(s, "")
}
