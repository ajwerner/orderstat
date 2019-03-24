package orderstat

import "fmt"

func (t *Tree) isBST() error {
	return t.root.isBST(t, nil, nil)
}

func (it *iterator) isBST(t *Tree, min, max Item) error {
	if it.node == nil {
		return nil
	}
	if min != nil && it.item.Less(min) {
		return fmt.Errorf("key %v < min %v", it.item, min)
	}
	if max != nil && max.Less(it.item) {
		return fmt.Errorf("key %v > max %v", it.item, max)
	}
	l := it.l(t)
	if l.node != nil && it.item.Less(l.item) {
		return fmt.Errorf("parent key %v < left child key %v", it.item, l.item)
	}
	r := it.r(t)
	if r.node != nil && r.item.Less(it.item) {
		return fmt.Errorf("parent key (%v) %v > right child key (%v)", it.np, it.item, r.np)
	}
	if err := l.isBST(t, min, it.item); err != nil {
		return err
	}
	if err := r.isBST(t, it.item, max); err != nil {
		return err
	}
	if lc, rc, ic := l.count(), r.count(), it.count(); ic != lc+rc+1 {
		return fmt.Errorf("count is not equal: %v + %v + 1 != %v", lc, rc, ic)
	}
	return nil
}
