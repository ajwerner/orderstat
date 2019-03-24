package benchmark

import (
	"math/rand"
	"testing"
)

type Int int

func (i Int) Less(item Item) bool {
	return i < item.(Int)
}

func makeData(n int) []Item {
	data := make([]Item, n)
	for i := 0; i < n; i++ {
		data[i] = Int(rand.Int())
	}
	return data
}

func BenchmarkGet(b *testing.B) {
	t := NewTree()
	data := makeData(b.N)
	for i := 0; i < b.N; i++ {
		t.ReplaceOrInsert(data[i])
	}
	perm := rand.Perm(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Get(data[perm[i]])
	}
}

func BenchmarkInsert(b *testing.B) {
	t := NewTree()
	data := makeData(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.ReplaceOrInsert(data[i])
	}
}
