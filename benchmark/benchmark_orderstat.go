// +build !btree

package benchmark

import "github.com/ajwerner/orderstat"

type Item = orderstat.Item
type ItemIterator = orderstat.ItemIterator

const Name = "orderstat"

func NewTree() Tree {
	return orderstat.NewTree()
}
