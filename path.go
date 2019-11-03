package redisgraph

import (
	"fmt"
	"strings"
)

type Path struct {
	Nodes []interface{}
	Edges []interface{}
}

func PathNew(nodes []interface{}, edges []interface{}) Path {
	return Path{
		Nodes: nodes,
		Edges: edges,
	}
}

func (p Path) GetNodes() []interface{} {
	return p.Nodes
}

func (p Path) GetEdges() []interface{} {
	return p.Edges
}

func (p Path) GetNode(index int) *Node {
	return p.Nodes[index].(*Node)
}

func (p Path) GetEdge(index int) *Edge{
	return p.Edges[index].(*Edge);
}

func (p Path) FirstNode() *Node {
	return p.GetNode(0)
}

func (p Path) LastNode() *Node {
	return p.GetNode(p.NodesCount() - 1)
}

func (p Path) NodesCount() int {
	return len(p.Nodes)
}

func (p Path) EdgeCount() int {
	return len(p.Edges)
}

func (p Path) Encode() string {
	s := []string{"<"}
	edgeCount := p.EdgeCount()
	for i := 0; i < edgeCount; i++ {
		var node = p.GetNode(i)
		s = append(s, "(" , fmt.Sprintf("%v", node.ID) , ")")
		var edge = p.GetEdge(i)
		if node.ID == edge.srcNodeID {
			s = append(s, "-[" , fmt.Sprintf("%v", edge.ID) , "]->")
		} else {
			s= append(s, "<-[" , fmt.Sprintf("%v", edge.ID) , "]-")
		}
	}
	s = append(s, "(" , fmt.Sprintf("%v", p.GetNode(edgeCount).ID) , ")")
	s = append(s, ">")

	return strings.Join(s, "")
}
