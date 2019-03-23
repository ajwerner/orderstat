package rankorder

import (
	"github.com/gonum/graph/encoding/dot"
	"github.com/gonum/graph/simple"
)

func (t *Tree) Plot() ([]byte, error) {
	g := simple.NewDirectedGraph(0, 0)
	t.root.addToGraph(t, g, Iterator{np: null})
	return dot.Marshal(g, "asdfasdf", "", "", false)
}

func (it Iterator) addToGraph(t *Tree, g *simple.DirectedGraph, p Iterator) {
	if it.node == nil {
		return
	}
	g.AddNode(simple.Node(it.np))
	if p.node != nil {
		g.HasEdgeFromTo(simple.Node(p.np), simple.Node(it.np))
	}
	it.l(t).addToGraph(t, g, it)
	it.r(t).addToGraph(t, g, it)
}
