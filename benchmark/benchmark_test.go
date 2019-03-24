package benchmark

import (
	"math/rand"
	"sort"
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
	b.Run(Name, func(b *testing.B) {
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
	})
}

func BenchmarkAscendRange(b *testing.B) {
	b.Run(Name, func(b *testing.B) {
		t := NewTree()
		data := makeData(b.N)
		for i := 0; i < b.N; i++ {
			t.ReplaceOrInsert(data[i])
		}
		pairs := make([]int, 0, 2*b.N)
		for i := 0; i < b.N; i++ {
			a := rand.Intn(b.N)
			b := rand.Intn(b.N - a)
			pairs = append(pairs, a, b)
		}
		sort.Slice(data, func(i, j int) bool {
			return data[i].(Int) < data[j].(Int)
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.AscendRange(data[pairs[2*i]], data[pairs[2*i+1]],
				func(item Item) bool { return true })
		}
	})
}

func BenchmarkInsert(b *testing.B) {
	b.Run(Name, func(b *testing.B) {
		t := NewTree()
		data := makeData(b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.ReplaceOrInsert(data[i])
		}
	})
}
