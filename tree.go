package rankorder

import (
	"fmt"
	"math"
)

type Item interface {
	Less(other interface{}) bool
}

type ItemIterator func(Item) (wantMore bool)

type Tree struct {
	root Iterator
	fp   Iterator
	list []node
}

func NewTree() *Tree {
	t := &Tree{}
	t.root.np = null
	t.fp.np = null
	return t
}

func (t *Tree) Ascend(f ItemIterator) {
	for it, ok := t.root.min(t); ok && f(it.k); it, ok = it.next(t) {
	}
}

func (t *Tree) AscendGreatorOrEqual(pivot Item, f ItemIterator) {
	var it Iterator
	ok := it.seek(t, pivot, seekGreatorOrEqual)
	for ; ok && f(it.k); it, ok = it.next(t) {
	}
}

func (t *Tree) AscendLessThan(pivot Item, f ItemIterator) {
	var limit Iterator
	if ok := limit.seek(t, pivot, seekLess); !ok {
		return
	}
	it, ok := t.root.min(t)
	for ; ok && f(it.k) && it.np != limit.np; it, ok = it.next(t) {
	}
}

func (t *Tree) AscendRange(greatorOrEqual, lessThan Item, f ItemIterator) {
	var limit Iterator
	if ok := limit.seek(t, lessThan, seekLess); !ok {
		return
	}
	var it Iterator
	ok := it.seek(t, greatorOrEqual, seekGreatorOrEqual)
	for ; ok && f(it.k) && it.np != limit.np; it, ok = it.next(t) {
	}
}

func (t *Tree) Delete(item Item) (replaced Item) {
	n := node{k: item, l: null, r: null, p: null}
	it := Iterator{node: &n}
	if !t.root.r(t).isRed() && !t.root.l(t).isRed() {
		t.root.setIsRed(true)
	}
	t.root, replaced = t.root.del(t, &it)
	t.root.setIsRed(false)
	return replaced
}

func (t *Tree) DeleteMin() (removed Item) {
	t.root, removed = t.root.delMin(t)
	return removed
}

func (t *Tree) DeleteMax() Item {
	panic("not implemented")
}

func (t *Tree) Descend(f ItemIterator) {
	for it, ok := t.root.max(t); ok && f(it.k); it, ok = it.prev(t) {
	}
}

func (t *Tree) DescendGreaterThan(pivot Item, f ItemIterator) {
	panic("not implemented")
}

func (t *Tree) DescendLessOrEqual(pivot Item, f ItemIterator) {
	panic("not implemented")
}

func (t *Tree) DescendRange(lessOrEqual, greaterThan Item, iterator ItemIterator) {

}

func (t *Tree) Get(key Item) Item {
	var it Iterator
	if it.seek(t, key, seekEqual) {
		return it.Item()
	}
	return nil
}

func (t *Tree) Has(key Item) bool {
	var it Iterator
	return it.seek(t, key, seekEqual)
}

func (t *Tree) Len() int {
	return int(t.root.count())
}

func (t *Tree) Max() Item {
	if it, ok := t.root.max(t); ok {
		return it.k
	}
	return nil
}

func (t *Tree) Min() Item {
	if it, ok := t.root.min(t); ok {
		return it.k
	}
	return nil
}

func (t *Tree) ReplaceOrInsert(item Item) (replaced Item) {
	new := t.alloc(item)
	t.root, replaced = t.root.add(t, new)
	t.root.setIsRed(false)
	return replaced
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

func (t *Tree) alloc(item Item) (it Iterator) {
	if t.fp.node == nil {
		t.realloc()
	}
	it = t.fp
	t.fp = it.r(t)
	*it.node = node{k: item, p: null, l: null, r: null, c: redMask}
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
	k Item
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
	return fmt.Sprintf("{%-2d %5.5v %.1v %-2d %-2d}", it.np, it.k, it.isRed(), it.node.l, it.node.r)
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

type seekMode int

const (
	_ seekMode = iota
	seekGreatorOrEqual
	seekEqual
	seekLess
)

func (it *Iterator) seek(t *Tree, item Item, mode seekMode) (ok bool) {
	*it = t.root
	for it.node != nil {
		switch {
		case item.Less(it.k):
			l := it.l(t)
			if l.node == nil {
				switch mode {
				case seekGreatorOrEqual:
					return true
				case seekLess:
					*it, ok = it.prev(t)
					return
				}

			}
			*it = l
		case it.k.Less(item):
			r := it.r(t)
			if r.node == nil {
				switch mode {
				case seekGreatorOrEqual:
					*it, ok = it.next(t)
					return
				case seekLess:
					return true
				}
			}
			*it = it.r(t)
		default:
			switch mode {
			case seekGreatorOrEqual, seekEqual:
				return true
			case seekLess:
				*it, ok = it.prev(t)
				return
			}
		}
	}
	return false
}

func (it *Iterator) Seek(t *Tree, item Item) (ok bool) {
	return it.seek(t, item, seekEqual)
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

func (it Iterator) next(t *Tree) (next Iterator, ok bool) {
	if it.hasRight() {
		return it.r(t).min(t)
	}
	p := it.p(t)
	for ; p.node != nil && p.node.l != it.np; it, p = p, p.p(t) {
	}
	return p, p.node != nil
}

func (it Iterator) prev(t *Tree) (next Iterator, ok bool) {
	if it.hasLeft() {
		return it.l(t).max(t)
	}
	p := it.p(t)
	for ; p.node != nil && p.node.r != it.np; it, p = p, p.p(t) {
	}
	return p, p.node != nil
}

func (it *Iterator) Item() Item { return it.k }

func (it Iterator) del(t *Tree, toDel *Iterator) (_ Iterator, replaced Item) {
	if it.node == nil {
		return Iterator{np: null}, nil
	}
	if less := toDel.k.Less(it.k); less {
		if l := it.l(t); !l.isRed() && !l.l(t).isRed() {
			it = it.moveRedLeft(t)
		}
		var l Iterator
		l, replaced = it.l(t).del(t, toDel)
		it.setLeft(l)
	} else {
		if it.l(t).isRed() {
			it = it.rotateRight(t)
		}
		if less = toDel.k.Less(it.k); !less && !it.k.Less(toDel.k) && !it.hasRight() {
			replaced = it.k
			t.free(it)
			return Iterator{np: null}, replaced
		}
		if r := it.r(t); !r.isRed() && !r.l(t).isRed() {
			it = it.moveRedRight(t)
		}
		if !toDel.k.Less(it.k) && !it.k.Less(toDel.k) {
			r := it.r(t)
			s, _ := r.min(t)
			replaced = it.k
			it.k = s.k
			var rem Item
			r, rem = r.delMin(t)
			it.setRight(r)
		} else {
			var r Iterator
			r, replaced = it.r(t).del(t, toDel)
			it.setRight(r)
		}
	}
	return it.fixUp(t), replaced
}

func (it Iterator) min(t *Tree) (Iterator, bool) {
	if it.node == nil {
		return it, false
	}
	for it.hasLeft() {
		it.init(t, it.node.l)
	}
	return it, true
}

func (it Iterator) max(t *Tree) (Iterator, bool) {
	if it.node == nil {
		return it, false
	}
	for it.hasRight() {
		it.init(t, it.node.r)
	}
	return it, true
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

func (it Iterator) delMin(t *Tree) (ret Iterator, removed Item) {
	if !it.hasLeft() {
		removed = it.k
		t.free(it)
		return Iterator{np: null}, removed
	}
	l := it.l(t)
	if !l.isRed() && !l.l(t).isRed() {
		it = it.moveRedLeft(t)
	}
	l, removed = it.l(t).delMin(t)
	it.setLeft(l)
	return it.fixUp(t), removed
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

func (it Iterator) add(t *Tree, toAdd Iterator) (ret Iterator, replaced Item) {
	if it.node == nil {
		toAdd.setIsRed(true)
		return toAdd.fixUp(t), nil
	}
	switch {
	case toAdd.k.Less(it.k):
		var l Iterator
		l, replaced = it.l(t).add(t, toAdd)
		it.setLeft(l)
	case it.k.Less(toAdd.k):
		var r Iterator
		r, replaced = it.r(t).add(t, toAdd)
		it.setRight(r)
	default:
		replaced = it.k
		it.k = toAdd.k
		t.free(toAdd)
		return it, replaced
	}

	return it.fixUp(t), replaced
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
