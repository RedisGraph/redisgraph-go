package redisgraph

import (
	"fmt"
	"strings"
)

// Node represents a node within a graph.
type Node struct {
	ID         uint64
	Label      string
	Alias      string
	Properties map[string]interface{}
	graph      *Graph
}

func NodeNew(label string, alias string, properties map[string]interface{}) *Node {

	p := properties
	if p == nil {
		p = make(map[string]interface{})
	}

	return &Node{
		Label:      label,
		Alias:      alias,
		Properties: p,
		graph:      nil,
	}
}

func (n *Node) SetProperty(key string, value interface{}) {
	n.Properties[key] = value
}

func (n Node) GetProperty(key string) interface{} {
	v, _ := n.Properties[key]
	return v
}

func (n Node) String() string {
	if len(n.Properties) == 0 {
		return "{}"
	}

	p := make([]string, 0, len(n.Properties))
	for k, v := range n.Properties {
		p = append(p, fmt.Sprintf("%s:%v", k, ToString(v)))
	}

	s := fmt.Sprintf("{%s}", strings.Join(p, ","))
	return s
}

// String makes Node satisfy the Stringer interface.
func (n Node) Encode() string {
	s := []string{"("}

	if n.Alias != "" {
		s = append(s, n.Alias)
	}

	if n.Label != "" {
		s = append(s, ":", n.Label)
	}

	if len(n.Properties) > 0 {
		p := make([]string, 0, len(n.Properties))
		for k, v := range n.Properties {
			p = append(p, fmt.Sprintf("%s:%v", k, ToString(v)))
		}

		s = append(s, "{")
		s = append(s, strings.Join(p, ","))
		s = append(s, "}")
	}

	s = append(s, ")")
	return strings.Join(s, "")
}
