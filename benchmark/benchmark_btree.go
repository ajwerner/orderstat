// +build btree

package benchmark

import "github.com/google/btree"

type Item = btree.Item
type ItemIterator = btree.ItemIterator

const Name = "btree"

func NewTree() Tree {
	return btree.New(2)
}
