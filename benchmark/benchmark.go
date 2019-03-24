package benchmark

type Tree interface {
	AscendGreaterOrEqual(pivot Item, iterator ItemIterator)
	AscendLessThan(pivot Item, iterator ItemIterator)
	AscendRange(greaterOrEqual, lessThan Item, iterator ItemIterator)
	Descend(iterator ItemIterator)
	DescendGreaterThan(pivot Item, iterator ItemIterator)
	DescendLessOrEqual(pivot Item, iterator ItemIterator)
	DescendRange(lessOrEqual, greaterThan Item, iterator ItemIterator)
	ReplaceOrInsert(item Item) Item
	Get(key Item) Item
	Delete(item Item) Item
}
