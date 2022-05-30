package redisgraph

import (
	"fmt"
	"strings"
)

// Node represents a node within a graph
type Node struct {
	ID         uint64
	Labels     []string
	Alias      string
	Properties map[string]interface{}
	graph      *Graph
}

// Create a new Node
func NodeNew(labels []string, alias string, properties map[string]interface{}) *Node {

	p := properties
	if p == nil {
		p = make(map[string]interface{})
	}

	return &Node{
		Labels:     labels,
		Alias:      alias,
		Properties: p,
		graph:      nil,
	}
}

// Asssign a new property to node
func (n *Node) SetProperty(key string, value interface{}) {
	n.Properties[key] = value
}

// Retrieves property from node
func (n Node) GetProperty(key string) interface{} {
	v, _ := n.Properties[key]
	return v
}

// Returns a string representation of a node
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

// String makes Node satisfy the Stringer interface
func (n Node) Encode() string {
	s := []string{"("}

	if n.Alias != "" {
		s = append(s, n.Alias)
	}

	for _, label := range n.Labels {
		s = append(s, ":", label)
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
