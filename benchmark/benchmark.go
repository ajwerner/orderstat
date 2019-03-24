package benchmark

type Tree interface {
	AscendGreaterOrEqual(pivot Item, iterator ItemIterator)
	AscendLessThan(pivot Item, iterator ItemIterator)
	AscendRange(greaterOrEqual, lessThan Item, iterator ItemIterator)
	ReplaceOrInsert(item Item) Item
	Get(key Item) Item
	Delete(item Item) Item
}
