package rankorder

import (
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

func CompareFloats(a, b interface{}) (cmp int) {
	if af, bf := a.(float64), b.(float64); af < bf {
		cmp = -1
	} else if bf < af {
		cmp = 1
	}
	return cmp
}

type Tree struct {
	cmp  Comparator
	root Iterator
	fp   Iterator
	list []node
}

func (t *Tree) Clone() *Tree {
	var cp Tree
	cp = *t
	newList := make([]node, len(t.list))
	copy(newList, t.list)
	cp.list = newList
	cp.fp.init(t, cp.fp.np)
	cp.root.init(t, cp.root.np)
	return &cp
}

func NewTree(cmp Comparator) *Tree {
	t := &Tree{cmp: cmp}
	t.root.np = null
	t.fp.np = null
	return t
}

func (t *Tree) Upsert(key, value interface{}) (replaced bool) {
	// fmt.Println("Upsert", key, t.root.np, t.root.node, "\n", t.list)
	new := t.alloc(key, value)
	// fmt.Println("alloc'ed", new, t.list)
	t.root, replaced = t.root.add(t, new)
	t.root.setIsRed(false)
	// t.root.cBST(t)
	// fmt.Println("Upsert", key, "\n", t)
	return replaced
}

func (t *Tree) Remove(key interface{}) (removed bool) {
	panic("not implemented")
}

func (t *Tree) realloc() {
	prevLen := len(t.list)
	var newList []node
	if prevLen > 0 {
		newList = make([]node, 2*prevLen)
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
	newList[len(newList)-1] = node{p: null, l: null, r: null}
	t.list = newList
	t.fp.init(t, pointer(prevLen))
	t.root.init(t, t.root.np)
}

func (t *Tree) alloc(k, v interface{}) (it Iterator) {
	if t.fp.node == nil {
		t.realloc()
	}
	it = t.fp
	t.fp = it.r(t)
	*it.node = node{k: k, v: v, p: null, l: null, r: null, c: redMask}
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
	n.c = ((n.c^redMask)&redMask | (n.c & countMask))
}

func (it Iterator) String() string {
	if it.node == nil {
		return "{null nil}"
	}
	return fmt.Sprintf("{%-2d %5.5v %5.5v %.1v %-2d %-2d}", it.np, it.k, it.v, it.isRed(), it.node.l, it.node.r)
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
	if !t.root.r(t).isRed() && !t.root.l(t).isRed() {
		t.root.setIsRed(true)
	}
	t.root, found = t.root.del(t, &it)
	t.root.setIsRed(false)
	return found
}

// // func (it *Iterator) Next() (ok bool)    { it.n = next(it.n); return it.cur != nil }
// // func (it *Iterator) Prev() (ok bool)    { it.n = prev(it.n); return it.cur != nil }

func (it *Iterator) Value() interface{} { return it.v }
func (it *Iterator) Key() interface{}   { return it.k }

func (it Iterator) del(t *Tree, toDel *Iterator) (_ Iterator, found bool) {
	if it.node == nil {
		return Iterator{np: null}, false
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
		if cmp = t.cmp(toDel.k, it.k); cmp == 0 && !it.hasRight() {
			t.free(it)
			return Iterator{np: null}, true
		}
		if r := it.r(t); !r.isRed() && !r.l(t).isRed() {
			it = it.moveRedRight(t)
		}
		if cmp = t.cmp(toDel.k, it.k); cmp == 0 {
			r := it.r(t)
			s := r.min(t)
			it.k = s.k
			it.v = s.v
			it.setRight(r.delMin(t))
			found = true
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
	it.setCount(l.count() + r.count() + 1)
	return it
}

func (it Iterator) delMin(t *Tree) (ret Iterator) {
	if !it.hasLeft() {
		t.free(it)
		return Iterator{np: null}
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

func (it Iterator) rotateRight(t *Tree) (ret Iterator) {
	if it.node == nil || !it.l(t).isRed() {
		panic("invalid rotate right")
	}
	x := it.l(t)
	it.setLeft(x.r(t))
	x.setRight(it)
	x.setIsRed(it.isRed())
	it.setIsRed(true)
	x.node.p = null
	x.setCount(x.l(t).count() + x.r(t).count() + 1)
	return x
}

func (it Iterator) rotateLeft(t *Tree) (ret Iterator) {
	// if it.node == nil || !it.r(t).isRed() {
	// 	panic(fmt.Sprintf("invalid rotate left %v %v", it, it.r(t)))
	// }
	x := it.r(t)
	it.setRight(x.l(t))
	x.setLeft(it)
	x.setIsRed(it.isRed())
	it.setIsRed(true)
	x.node.p = null
	x.setCount(x.l(t).count() + x.r(t).count() + 1)
	return x
}

func (it Iterator) moveRedLeft(t *Tree) (ret Iterator) {
	// if it.node == nil || !(it.isRed() && !it.l(t).isRed() && !it.l(t).l(t).isRed()) {
	// 	panic(fmt.Sprintf("invalid moveRedLeft %v %v %v", it, it.l(t), it.l(t).l(t)))
	// }
	it.colorFlip(t)
	if r := it.r(t); r.l(t).isRed() {
		it.setRight(r.rotateRight(t))
		it = it.rotateLeft(t)
		it.colorFlip(t)
	}
	return it
}

func (it Iterator) moveRedRight(t *Tree) (ret Iterator) {
	// if it.node == nil || !(it.isRed() && !it.r(t).isRed() && !it.r(t).l(t).isRed()) {
	// 	panic("invalid moveRedLeft")
	// }
	it.colorFlip(t)
	if l := it.l(t); l.l(t).isRed() {
		it = it.rotateRight(t)
		it.colorFlip(t)
	}
	return it
}

func (it Iterator) add(t *Tree, toAdd Iterator) (ret Iterator, replaced bool) {
	if it.node == nil {
		toAdd.setIsRed(true)
		return toAdd.fixUp(t), false
	}

	cmp := t.cmp(toAdd.k, it.k)
	switch {
	case cmp < 0:
		var l Iterator
		l, replaced = it.l(t).add(t, toAdd)
		it.setLeft(l)
	case cmp == 0:
		it.k, it.v = toAdd.k, toAdd.v
		t.free(toAdd)
		return it, true
	case cmp > 0:
		var r Iterator
		r, replaced = it.r(t).add(t, toAdd)
		it.setRight(r)
	}

	return it.fixUp(t), replaced
}

func (it *Iterator) isBST(t *Tree, min, max interface{}) error {
	// fmt.Println("isBST ", it.np)
	if it.node == nil {
		return nil
	}
	// fmt.Println("isBST not nil", it)
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
	if r.node != nil {
		// fmt.Println("aqui", r, it, t.list)
		if t.cmp(it.k, r.k) > 0 {
			return fmt.Errorf("parent key (%v) %v > right child key (%v) %v", it.np, it.k, r.np, r.v)
		}
	}
	// TODO: parent check
	// fmt.Println("isBST left", it.np, l.np)
	if err := l.isBST(t, min, it.k); err != nil {
		return err
	}
	// fmt.Println("isBST right", it.np, r.np)
	if err := r.isBST(t, it.k, max); err != nil {
		return err
	}
	return nil
}

// func (t *Tree) String() string {
// 	_, bufs := t.root.print(t, 0, 0, nil)
// 	if len(bufs) == 0 {
// 		return ""
// 	}
// 	buf := bufs[0]
// 	for _, b := range bufs[1:] {
// 		buf.WriteString("\n")
// 		buf.ReadFrom(b)
// 	}
// 	buf.WriteString("\n")
// 	return buf.String()
// }

// func (it Iterator) print(t *Tree, col, depth int, bufs []*bytes.Buffer) (int, []*bytes.Buffer) {
// 	if it.node == nil {
// 		return col, bufs
// 	}
// 	if len(bufs) < depth+1 {
// 		bufs = append(bufs, &bytes.Buffer{})
// 	}
// 	col, bufs = it.l(t).print(t, col, depth+1, bufs)
// 	buf := bufs[depth]
// 	fmt.Fprintf(buf, "%s%v", strings.Repeat(" ", col), it.String())
// 	_, bufs = it.r(t).print(t, col-8, depth+1, bufs)
// 	return col + 16, bufs
// }
