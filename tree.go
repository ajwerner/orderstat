package rankorder

import (
	"fmt"
	"math"
)

////////////////////////////////////////////////////////////////////////////////
// Public API
////////////////////////////////////////////////////////////////////////////////

// Item represents a single object in the tree.
type Item interface {

	// Less tests whether the current item is less than the given argument.
	//
	// This must provide a strict weak ordering.
	// If !a.Less(b) && !b.Less(a), we treat this to mean a == b (i.e. we can only
	// hold one of either a or b in the tree).
	Less(other Item) bool
}

// ItemIterator allows callers of Ascend* to iterate in-order over
// portions of the tree. When this function returns false, iteration will
// stop and the associated Ascend* function will immediately return.
type ItemIterator func(Item) (wantMore bool)

// Tree stores Item instances in an ordered structure, allowing easy removal,
// and iteration.
//
// Write operations are not safe for concurrent mutation by multiple goroutines,
// but Read operations are.
type Tree struct {
	root iterator
	fp   iterator
	list []node
}

// NewTree creates a new Tree.
func NewTree() *Tree {
	t := &Tree{}
	t.root.np = null
	t.fp.np = null
	return t
}

// Ascend calls the iterator for every value in the tree within the range
// [first, last], until iterator returns false.
func (t *Tree) Ascend(f ItemIterator) {
	for it, ok := t.root.min(t); ok && f(it.item); it, ok = it.next(t) {
	}
}

// AscendGreaterOrEqual calls the iterator for every value in the tree within
// the range [pivot, last], until iterator returns false.
func (t *Tree) AscendGreaterOrEqual(pivot Item, f ItemIterator) {
	var it iterator
	ok := it.seek(t, pivot, seekGTE)
	for ; ok && f(it.item); it, ok = it.next(t) {
	}
}

// AscendLessThan calls the iterator for every value in the tree within the range
// [first, pivot), until iterator returns false.
func (t *Tree) AscendLessThan(pivot Item, f ItemIterator) {
	var limit iterator
	if ok := limit.seek(t, pivot, seekLT); !ok {
		return
	}
	it, ok := t.root.min(t)
	for ; ok && f(it.item) && it.np != limit.np; it, ok = it.next(t) {
	}
}

// AscendLessThan calls the iterator for every value in the tree within the range
// [first, pivot), until iterator returns false.
func (t *Tree) AscendRange(greatorOrEqual, lessThan Item, f ItemIterator) {
	var limit iterator
	if ok := limit.seek(t, lessThan, seekLT); !ok {
		return
	}
	var it iterator
	ok := it.seek(t, greatorOrEqual, seekGTE)
	for ; ok && f(it.item) && it.np != limit.np; it, ok = it.next(t) {
	}
}

// Delete removes an item equal to the passed in item from the tree, returning
// it. If no such item exists, returns nil.
func (t *Tree) Delete(item Item) (replaced Item) {
	n := node{item: item, l: null, r: null, p: null}
	it := iterator{node: &n}
	if !t.root.r(t).isRed() && !t.root.l(t).isRed() {
		t.root.setIsRed(true)
	}
	t.root, replaced = t.root.del(t, &it)
	t.root.setIsRed(false)
	return replaced
}

// DeleteMin removes the smallest item in the tree and returns it.
// If no such item exists, returns nil.
func (t *Tree) DeleteMin() (removed Item) {
	t.root, removed = t.root.delMin(t)
	return removed
}

// DeleteMax removes the largest item in the tree and returns it.
// If no such item exists, returns nil.
func (t *Tree) DeleteMax() Item {
	return t.Delete(t.Max())
}

// Descend calls the iterator for every value in the tree within the range
// [last, first], until iterator returns false.
func (t *Tree) Descend(f ItemIterator) {
	for it, ok := t.root.max(t); ok && f(it.item); it, ok = it.prev(t) {
	}
}

// DescendGreaterThan calls the iterator for every value in the tree within
// the range (pivot, last], until iterator returns false.
func (t *Tree) DescendGreaterThan(pivot Item, f ItemIterator) {
	var limit iterator
	if ok := limit.seek(t, pivot, seekGT); !ok {
		return
	}
	it, ok := t.root.max(t)
	for ; ok && f(it.item) && it.np != limit.np; it, ok = it.prev(t) {
	}
}

// DescendLessOrEqual calls the iterator for every value in the tree within the
// range [pivot, first], until iterator returns false.
func (t *Tree) DescendLessOrEqual(pivot Item, f ItemIterator) {
	var it iterator
	ok := it.seek(t, pivot, seekLTE)
	for ; ok && f(it.item); it, ok = it.prev(t) {
	}
}

// DescendRange calls the iterator for every value in the tree within the range
// [lessOrEqual, greaterThan), until iterator returns false.
func (t *Tree) DescendRange(lessOrEqual, greaterThan Item, f ItemIterator) {
	var limit iterator
	if ok := limit.seek(t, greaterThan, seekGT); !ok {
		return
	}
	var it iterator
	ok := it.seek(t, lessOrEqual, seekLTE)
	for ; ok && f(it.item) && it.np != limit.np; it, ok = it.prev(t) {
	}
}

func (t *Tree) Get(key Item) Item {
	var it iterator
	if it.seek(t, key, seekEQ) {
		return it.item
	}
	return nil
}

func (t *Tree) Has(key Item) bool {
	var it iterator
	return it.seek(t, key, seekEQ)
}

func (t *Tree) Len() int {
	return int(t.root.count())
}

func (t *Tree) Max() Item {
	if it, ok := t.root.max(t); ok {
		return it.item
	}
	return nil
}

func (t *Tree) Min() Item {
	if it, ok := t.root.min(t); ok {
		return it.item
	}
	return nil
}

func (t *Tree) ReplaceOrInsert(item Item) (replaced Item) {
	new := t.alloc(item)
	t.root, replaced = t.root.add(t, new)
	t.root.setIsRed(false)
	return replaced
}

////////////////////////////////////////////////////////////////////////////////
// Memory management
////////////////////////////////////////////////////////////////////////////////

type pointer uint32

const null pointer = math.MaxUint32

func (p pointer) n(t *Tree) *node {
	if p == null {
		return nil
	}
	return &t.list[int(p)]
}

const redMask uint32 = 1 << 31
const countMask uint32 = ^redMask

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

func (t *Tree) alloc(item Item) (it iterator) {
	if t.fp.node == nil {
		t.realloc()
	}
	it = t.fp
	t.fp = it.r(t)
	*it.node = node{item: item, p: null, l: null, r: null, c: redMask}
	return it
}

func (t *Tree) free(it iterator) {
	*it.node = node{l: null, r: null, p: null}
	it.setRight(t.fp)
	t.fp = it
}

////////////////////////////////////////////////////////////////////////////////
// node
////////////////////////////////////////////////////////////////////////////////

type node struct {
	item Item
	l    pointer
	r    pointer
	p    pointer
	c    uint32
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

func (n *node) hasLeft() bool {
	return n != nil && n.l != null
}
func (n *node) hasRight() bool {
	return n != nil && n.r != null
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

func (it iterator) String() string {
	if it.node == nil {
		return "{null nil}"
	}
	return fmt.Sprintf("{%-2d %5.5v %.1v %-2d %-2d}", it.np, it.item, it.isRed(), it.node.l, it.node.r)
}

////////////////////////////////////////////////////////////////////////////////
// iterator
////////////////////////////////////////////////////////////////////////////////

type iterator struct {
	*node
	np pointer
}

func (it *iterator) init(t *Tree, p pointer) bool {
	*it = iterator{np: p, node: p.n(t)}
	return it.node != nil
}

func (it iterator) p(t *Tree) (p iterator) {
	if it.node != nil {
		p.init(t, it.node.p)
	}
	return p
}

func (it iterator) l(t *Tree) (l iterator) {
	if it.node != nil {
		l.init(t, it.node.l)
	}
	return l
}

func (it iterator) r(t *Tree) (r iterator) {
	if it.node != nil {
		r.init(t, it.node.r)
	}
	return r
}

func (it iterator) min(t *Tree) (iterator, bool) {
	if it.node == nil {
		return it, false
	}
	for it.hasLeft() {
		it.init(t, it.node.l)
	}
	return it, true
}

func (it iterator) max(t *Tree) (iterator, bool) {
	if it.node == nil {
		return it, false
	}
	for it.hasRight() {
		it.init(t, it.node.r)
	}
	return it, true
}

func (it iterator) next(t *Tree) (next iterator, ok bool) {
	if it.hasRight() {
		return it.r(t).min(t)
	}
	p := it.p(t)
	for ; p.node != nil && p.node.l != it.np; it, p = p, p.p(t) {
	}
	return p, p.node != nil
}

func (it iterator) prev(t *Tree) (next iterator, ok bool) {
	if it.hasLeft() {
		return it.l(t).max(t)
	}
	p := it.p(t)
	for ; p.node != nil && p.node.r != it.np; it, p = p, p.p(t) {
	}
	return p, p.node != nil
}

func (it iterator) setRight(r iterator) {
	if it.node == nil {
		return
	}
	it.node.r = r.np
	if r.node != nil {
		r.node.p = it.np
	}
}

func (it iterator) setLeft(l iterator) {
	if it.node == nil {
		return
	}
	it.node.l = l.np
	if l.node != nil {
		l.node.p = it.np
	}
}

func (it iterator) fixUp(t *Tree) (ret iterator) {
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

func (it iterator) add(t *Tree, toAdd iterator) (ret iterator, replaced Item) {
	if it.node == nil {
		toAdd.setIsRed(true)
		return toAdd.fixUp(t), nil
	}
	switch {
	case toAdd.item.Less(it.item):
		var l iterator
		l, replaced = it.l(t).add(t, toAdd)
		it.setLeft(l)
	case it.item.Less(toAdd.item):
		var r iterator
		r, replaced = it.r(t).add(t, toAdd)
		it.setRight(r)
	default:
		replaced = it.item
		it.item = toAdd.item
		t.free(toAdd)
		return it, replaced
	}

	return it.fixUp(t), replaced
}

func (it iterator) del(t *Tree, toDel *iterator) (_ iterator, replaced Item) {
	if it.node == nil {
		return iterator{np: null}, nil
	}
	if less := toDel.item.Less(it.item); less {
		if l := it.l(t); !l.isRed() && !l.l(t).isRed() {
			it = it.moveRedLeft(t)
		}
		var l iterator
		l, replaced = it.l(t).del(t, toDel)
		it.setLeft(l)
	} else {
		if it.l(t).isRed() {
			it = it.rotateRight(t)
		}
		if less = toDel.item.Less(it.item); !less && !it.item.Less(toDel.item) && !it.hasRight() {
			replaced = it.item
			t.free(it)
			return iterator{np: null}, replaced
		}
		if r := it.r(t); !r.isRed() && !r.l(t).isRed() {
			it = it.moveRedRight(t)
		}
		if !toDel.item.Less(it.item) && !it.item.Less(toDel.item) {
			r := it.r(t)
			replaced = it.item
			r, it.item = r.delMin(t)
			it.setRight(r)
		} else {
			var r iterator
			r, replaced = it.r(t).del(t, toDel)
			it.setRight(r)
		}
	}
	return it.fixUp(t), replaced
}

func (it iterator) delMin(t *Tree) (ret iterator, removed Item) {
	if !it.hasLeft() {
		removed = it.item
		t.free(it)
		return iterator{np: null}, removed
	}
	l := it.l(t)
	if !l.isRed() && !l.l(t).isRed() {
		it = it.moveRedLeft(t)
	}
	l, removed = it.l(t).delMin(t)
	it.setLeft(l)
	return it.fixUp(t), removed
}

func colorFlip(it, r, l *iterator) {
	it.flipRed()
	r.flipRed()
	l.flipRed()
}

func (it iterator) colorFlip(t *Tree) {
	r, l := it.r(t), it.l(t)
	colorFlip(&it, &r, &l)
}

func (it iterator) rotateRight(t *Tree) (ret iterator) {
	if it.node == nil || !it.l(t).isRed() {
		panic("invalid rotate right")
	}
	x := it.l(t)
	it.setLeft(x.r(t))
	x.setRight(it)
	x.setIsRed(it.isRed())
	it.setIsRed(true)
	x.node.p = null
	x.setCount(it.count())
	it.setCount(it.l(t).count() + it.r(t).count() + 1)
	return x
}

func (it iterator) rotateLeft(t *Tree) (ret iterator) {
	// if it.node == nil || !it.r(t).isRed() {
	// 	panic(fmt.Sprintf("invalid rotate left %v %v", it, it.r(t)))
	// }
	x := it.r(t)
	it.setRight(x.l(t))
	x.setLeft(it)
	x.setIsRed(it.isRed())
	it.setIsRed(true)
	x.node.p = null
	x.setCount(it.count())
	it.setCount(it.l(t).count() + it.r(t).count() + 1)
	return x
}

func (it iterator) moveRedLeft(t *Tree) (ret iterator) {
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

func (it iterator) moveRedRight(t *Tree) (ret iterator) {
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

////////////////////////////////////////////////////////////////////////////////
// Seek
////////////////////////////////////////////////////////////////////////////////

type seekMode int

const (
	_ seekMode = iota
	seekGT
	seekGTE
	seekLT
	seekLTE
	seekEQ
)

func (it *iterator) seek(t *Tree, item Item, mode seekMode) (ok bool) {
	*it = t.root
	for it.node != nil {
		switch {
		case item.Less(it.item):
			l := it.l(t)
			if l.node == nil {
				switch mode {
				case seekGTE, seekGT:
					return true
				case seekLTE, seekLT:
					*it, ok = it.prev(t)
					return
				}
			}
			*it = l
		case it.item.Less(item):
			r := it.r(t)
			if r.node == nil {
				switch mode {
				case seekGTE, seekGT:
					*it, ok = it.next(t)
					return
				case seekLTE, seekLT:
					return true
				}
			}
			*it = it.r(t)
		default:
			switch mode {
			case seekGTE, seekEQ, seekLTE:
				return true
			case seekGT:
				*it, ok = it.next(t)
				return
			case seekLT:
				*it, ok = it.prev(t)
				return
			}
		}
	}
	return false
}
