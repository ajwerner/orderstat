// +build !btree

package benchmark

import "github.com/ajwerner/rankorder"

type Item = rankorder.Item
type ItemIterator = rankorder.ItemIterator

const Name = "orderstat"

func NewTree() Tree {
	return rankorder.NewTree()
}
