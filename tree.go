package rankorder

import (
	"bytes"
	"fmt"
	"math"
)

func LessComparator(less func(a, b interface{}) (less bool)) Comparator {
	return func(a, b interface{}) int {
		if less(a, b) {
			return -1
		}
		if less(b, a) {
			return 1
		}
		return 0
	}
}

// Comparator returns 0 if a == b, return a negative int if a < b and a
// positive int if a > b.
type Comparator func(a, b interface{}) int

type Tree struct {
	cmp  Comparator
	root Iterator
	fp   Iterator
	list []node
}

func NewTree(cmp Comparator) *Tree {
	t := &Tree{cmp: cmp}
	t.root.np = null
	t.fp.np = null
	return t
}

func (t *Tree) Upsert(key, value interface{}) (replaced bool) {
	fmt.Println("Upsert", key)
	new := t.alloc(key, value)
	t.root, replaced = t.root.add(t, new)
	t.root.cBST(t)
	return replaced
}

func (t *Tree) Remove(key interface{}) (removed bool) {
	panic("not implemented")
}

func (t *Tree) realloc() {
	prevLen := len(t.list)
	var newList []node
	if prevLen > 0 {
		newList := make([]node, 2*prevLen)
		copy(newList, t.list)
	} else {
		const defaultSize = 16
		newList = make([]node, defaultSize)
	}
	for i := prevLen + 1; i < len(newList); i++ {
		newList[i-1] = node{
			p: null,
			l: null,
			r: pointer(i),
		}
	}
	t.fp = Iterator{
		node: &newList[prevLen],
		np:   pointer(prevLen),
	}
	t.list = newList
}

func (t *Tree) alloc(k, v interface{}) (it Iterator) {
	if t.fp.node == nil {
		t.realloc()
	}
	it = t.fp
	t.fp = it.r(t)
	*it.node = node{k: k, v: v, p: null, l: null, r: null}
	return it
}

func (t *Tree) free(it Iterator) {
	*it.node = node{l: null, r: null, p: null}
	it.setRight(t.fp)
	t.fp = it
}

type pointer uint32

const null pointer = math.MaxUint32

func (p pointer) n(t *Tree) *node {
	if p == null { // maybe hubris
		return nil

	}
	return &t.list[int(p)]
}

const redMask uint32 = 1 << 31
const countMask uint32 = ^redMask

type node struct {
	k interface{} // key
	v interface{} // value
	l pointer
	r pointer
	p pointer
	c uint32
}

func (n *node) isRed() bool {
	return n != nil && n.c&redMask != 0
}

func (n *node) flipRed() {
	if n == nil {
		return
	}
	n.c = n.c | ((n.c ^ redMask) | countMask)
}

func (n *node) setIsRed(to bool) {
	if n == nil {
		return
	}
	if to {
		n.c = n.c | redMask
	} else {
		n.c = n.c & countMask
	}
}

func (n *node) count() uint32 {
	if n == nil {
		return 0
	}
	return n.c & countMask
}

func (n *node) setCount(to uint32) {
	n.c = (n.c & redMask) | to
}

func (n *node) hasLeft() bool  { return n.l != null }
func (n *node) hasRight() bool { return n.r != null }

type Iterator struct {
	*node
	np pointer
}

func (it *Iterator) init(t *Tree, p pointer) bool {
	*it = Iterator{np: p, node: p.n(t)}
	return it.node != nil
}

func (it Iterator) p(t *Tree) (p Iterator) {
	if it.node != nil {
		p.init(t, it.node.p)
	}
	return p
}

func (it Iterator) l(t *Tree) (l Iterator) {
	if it.node != nil {
		l.init(t, it.node.l)
	}
	return l
}

func (it Iterator) r(t *Tree) (r Iterator) {
	if it.node != nil {
		r.init(t, it.node.r)
	}
	return r
}

func (it *Iterator) Seek(t *Tree, key interface{}) (ok bool) {
	*it = t.root
	for it.node != nil {
		cmp := t.cmp(key, it.k)
		switch {
		case cmp < 0:
			*it = it.l(t)
		case cmp == 0:
			return true
		case cmp > 0:
			*it = it.r(t)
		}
	}
	return false
}

func (it *Iterator) SeekCeil(t *Tree, key interface{}) (ok bool) {
	panic("not implemented")
}

func (it *Iterator) SeekFloor(t *Tree, key interface{}) (ok bool) {
	panic("not implemented")
}

func (t *Iterator) SeekRank(it *Tree, r int) (ok bool) {
	panic("not implemented")
}

func (t *Tree) Delete(k interface{}) (found bool) {
	n := node{k: k, l: null, r: null, p: null}
	it := Iterator{node: &n}
	_, found = t.root.del(t, &it)
	return found
}

// // func (it *Iterator) Next() (ok bool)    { it.n = next(it.n); return it.cur != nil }
// // func (it *Iterator) Prev() (ok bool)    { it.n = prev(it.n); return it.cur != nil }

func (it *Iterator) Value() interface{} { return it.v }
func (it *Iterator) Key() interface{}   { return it.k }

func (it Iterator) del(t *Tree, toDel *Iterator) (_ Iterator, found bool) {
	if it.node == nil {
		return it, false
	}
	if cmp := t.cmp(toDel.k, it.k); cmp < 0 {
		if l := it.l(t); !l.isRed() && !l.l(t).isRed() {
			it = it.moveRedLeft(t)
		}
		var l Iterator
		l, found = it.l(t).del(t, toDel)
		it.setLeft(l)
	} else {
		if it.l(t).isRed() {
			it = it.rotateRight(t)
		}
		if cmp == 0 && !it.hasRight() {
			t.free(it)
			return Iterator{np: null}, true
		}
		if r := it.r(t); !r.isRed() && !r.l(t).isRed() {
			it = it.moveRedRight(t)
		}
		if it.np == toDel.np {
			r := it.r(t)
			s := r.min(t)
			it.v = s.v
			it.setRight(r.delMin(t))
		} else {
			var r Iterator
			r, found = it.r(t).del(t, toDel)
			it.setRight(r)
		}
	}
	return it.fixUp(t), found
}

func (it Iterator) min(t *Tree) Iterator {
	for it.hasLeft() {
		it.init(t, it.node.l)
	}
	return it
}

func (it Iterator) max(t *Tree) Iterator {
	for it.hasRight() {
		it.init(t, it.node.r)
	}
	return it
}

func (it Iterator) fixUp(t *Tree) (ret Iterator) {
	defer it.cBST(t)(&ret)
	if it.r(t).isRed() {
		it = it.rotateLeft(t)
	}
	if l := it.l(t); l.isRed() && l.l(t).isRed() {
		it = it.rotateRight(t)
	}
	l, r := it.l(t), it.r(t)
	if l.isRed() && r.isRed() {
		colorFlip(&it, &l, &r)
	}
	it.setCount(l.count() + r.count())
	return it
}

func (it Iterator) cBST(t *Tree) func(it *Iterator) {
	// it.checkBST(t)
	fmt.Println(t.String())
	return func(it *Iterator) { it.checkBST(t) }
}

func (it Iterator) checkBST(t *Tree) {
	if err := it.isBST(t, nil, nil); err != nil {
		panic(err)
	}
}

func (it Iterator) delMin(t *Tree) (ret Iterator) {
	defer it.cBST(t)(&ret)
	if !it.hasLeft() {
		t.free(it)
		return Iterator{}
	}
	if l := it.l(t); !l.isRed() && !l.l(t).isRed() {
		it = it.moveRedLeft(t)
	}
	it.setLeft(it.l(t).delMin(t))
	return it.fixUp(t)
}

func (it Iterator) setRight(r Iterator) {
	if it.node == nil {
		return
	}
	it.node.r = r.np
	if r.node != nil {
		r.node.p = it.np
	}
}

func (it Iterator) setLeft(l Iterator) {
	if it.node == nil {
		return
	}
	it.node.l = l.np
	if l.node != nil {
		l.node.p = it.np
	}
}

func colorFlip(it, r, l *Iterator) {
	it.flipRed()
	r.flipRed()
	l.flipRed()
}

func (it Iterator) colorFlip(t *Tree) {
	r, l := it.r(t), it.l(t)
	colorFlip(&it, &r, &l)
}

func (it Iterator) rotateLeft(t *Tree) (ret Iterator) {
	defer it.cBST(t)(&ret)
	x := it.r(t)
	it.setRight(x.l(t))
	x.setLeft(it)
	x.setIsRed(it.l(t).isRed())
	it.l(t).setIsRed(true)
	x.node.p = null
	return x
}

func (it Iterator) rotateRight(t *Tree) (ret Iterator) {
	defer it.cBST(t)(&ret)
	x := it.l(t)
	it.setLeft(x.r(t))
	x.setRight(it)
	r := it.r(t)
	x.setIsRed(r.isRed())
	r.setIsRed(true)
	x.node.p = null
	return x
}

func (it Iterator) moveRedLeft(t *Tree) (ret Iterator) {
	defer it.cBST(t)(&ret)
	it.colorFlip(t)
	if r := it.r(t); r.l(t).isRed() {
		r = r.rotateRight(t)
		it.setRight(r)
		it = it.rotateLeft(t)
		it.colorFlip(t)
	}
	return it
}

func (it Iterator) moveRedRight(t *Tree) (ret Iterator) {
	defer it.cBST(t)(&ret)
	it.colorFlip(t)
	if l := it.l(t); l.l(t).isRed() {
		it = it.rotateRight(t)
		it.colorFlip(t)
	}
	return it
}

func (it Iterator) add(t *Tree, toAdd Iterator) (ret Iterator, replaced bool) {
	fmt.Println("asdf")
	defer it.cBST(t)(&ret)
	if it.node == nil {
		toAdd.setIsRed(true)
		return toAdd.fixUp(t), false
	}

	cmp := t.cmp(toAdd.k, it.k)
	fmt.Println("add ", it.k, toAdd.k, it.r(t).node, cmp)
	switch {
	case cmp < 0:
		var l Iterator
		l, replaced = it.l(t).add(t, toAdd)
		it.setLeft(l)
		fmt.Println("left ", l.k, l.node.l, l.node.r, "it", it.k, it.node.l, it.node.r)
	case cmp == 0:
		it.k, it.v = toAdd.k, toAdd.v
		t.free(toAdd)
		return it, true
	case cmp > 0:
		var r Iterator
		r, replaced = it.r(t).add(t, toAdd)
		it.setRight(r)
		fmt.Println("right ", r.k, r.node.l, r.node.r, "it", it.k, it.node.l, it.node.r)
	}
	return it.fixUp(t), replaced
}

func (it *Iterator) isBST(t *Tree, min, max interface{}) error {
	if it.node == nil {
		return nil
	}
	if min != nil && t.cmp(it.k, min) < 0 {
		return fmt.Errorf("key %v < min %v", it.k, min)
	}
	if max != nil && t.cmp(it.k, max) > 0 {
		return fmt.Errorf("key %v > max %v", it.k, max)
	}
	l := it.l(t)
	//fmt.Println(l.node, it.node)
	if l.node != nil && t.cmp(it.k, l.k) < 0 {
		return fmt.Errorf("parent key %v < left child key %v", it.k, l.k)
	}
	r := it.r(t)
	if r.node != nil && t.cmp(it.k, r.k) > 0 {
		return fmt.Errorf("parent key (%v) %v > right child key (%v) %v", it.np, it.k, r.np, r.v)
	}
	// TODO: parent check
	if err := l.isBST(t, min, it.k); err != nil {
		return err
	}
	if err := r.isBST(t, it.k, max); err != nil {
		return err
	}
	return nil
}

func (t *Tree) String() string {
	nodes := t.root.print(t, 0, nil)
	buf := &bytes.Buffer{}
	for _, nodes := range nodes {
		for i, n := range nodes {
			fmt.Fprintf(buf, "{%v %v %v}", n.k, n.v, n.isRed())
			if i > 0 {
				buf.WriteString(" ")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func (it Iterator) print(t *Tree, depth int, nodes [][]*node) [][]*node {
	if it.node == nil {
		return nodes
	}
	if len(nodes) < depth+1 {
		nodes = append(nodes, []*node{})
	}
	nodes[depth] = append(nodes[depth], it.node)
	nodes = it.l(t).print(t, depth+1, nodes)
	nodes = it.r(t).print(t, depth+1, nodes)
	return nodes
}
