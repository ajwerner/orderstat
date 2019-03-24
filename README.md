# orderstat

Order statistic tree implemented as a LLRB tree.

This package provides an in-memory Left-Leaning Red-Black tree implementation for Go, useful as an ordered, mutable data structure.

The API is based off of the wonderful http://godoc.org/github.com/petar/GoLLRB/llrb, and is meant to allow btree to act as a drop-in replacement for gollrb trees. In addition to that API it exposes Rank and Select methods.

See http://godoc.org/github.com/ajwerner/orderstat for documentation.
