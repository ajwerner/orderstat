package benchmark

import (
	"math/rand"
	"sort"
	"strconv"
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
		for _, l := range []int{1, 10, 100, 1000} {
			b.Run(strconv.Itoa(l), func(b *testing.B) {
				N := b.N + l
				t := NewTree()
				data := makeData(N)
				for i := 0; i < N; i++ {
					t.ReplaceOrInsert(data[i])
				}
				pairs := make([]Int, 0, 2*b.N)
				for i := 0; i < b.N; i++ {
					a := rand.Intn(N - l)
					b := a + l
					pairs = append(pairs, Int(a), Int(b))
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
