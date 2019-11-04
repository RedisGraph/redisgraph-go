package redisgraph

import (
	"fmt"
	"strings"
)

type Path struct {
	Nodes []*Node
	Edges []*Edge
}

func PathNew(nodes []interface{}, edges []interface{}) Path {
	Nodes := make([]*Node, len(nodes))
	for i := 0; i < len(nodes); i++ {
		Nodes[i] = nodes[i].(*Node)
	}
	Edges := make([]*Edge, len(edges))
	for i := 0; i < len(edges); i++ {
		Edges[i] = edges[i].(*Edge)
	}
	
	return Path{	
		Edges : Edges,
		Nodes : Nodes,
	}
}

func (p Path) GetNodes() []*Node {
	return p.Nodes
}

func (p Path) GetEdges() []*Edge {
	return p.Edges
}

func (p Path) GetNode(index int) *Node {
	return p.Nodes[index]
}

func (p Path) GetEdge(index int) *Edge{
	return p.Edges[index]
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

func (p Path) String() string {
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
