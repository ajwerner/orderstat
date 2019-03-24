package orderstat

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type keyValue struct {
	k string
	v string
}

func (kv keyValue) Less(other Item) bool {
	return kv.k < other.(keyValue).k
}

func kv(k, v string) keyValue { return keyValue{k: k, v: v} }

type intItem int

func (ii intItem) Less(other Item) bool {
	return ii < other.(intItem)
}

func TestAscendAndDescend(t *testing.T) {
	tr := NewTree()
	var items []int
	const N = 10000
	for i := 0; i < N; i++ {
		items = append(items, i)
	}
	shuf := rand.Perm(N)
	for i := 0; i < N; i++ {
		assert.Nil(t, tr.ReplaceOrInsert(intItem(items[shuf[i]])))
		assert.Equal(t, tr.Len(), i+1)
	}
	seen := 0
	action := func() { seen++ }
	expect := func(item Item) bool {
		assert.Equal(t, int(item.(intItem)), items[seen])
		action()
		return true
	}
	tr.Ascend(expect)
	seen = rand.Intn(N)
	tr.AscendGreaterOrEqual(intItem(items[seen]), expect)
	max := rand.Intn(N)
	seen = 0
	tr.AscendLessThan(intItem(items[max]), expect)
	assert.Equal(t, max, seen)
	tr.Delete(intItem(max))
	seen = 0
	tr.AscendLessThan(intItem(items[max]), expect)
	assert.Equal(t, seen, max)
	tr.ReplaceOrInsert(intItem(max))
	seen = rand.Intn(max)
	tr.AscendRange(intItem(seen), intItem(max), expect)
	assert.Equal(t, seen, max)
	seen = len(items) - 1
	action = func() { seen-- }
	tr.Descend(expect)
	min := rand.Intn(N)
	seen = len(items) - 1
	tr.DescendGreaterThan(intItem(min), expect)
	assert.Equal(t, seen, min)
}

func TestTree(t *testing.T) {
	tr := NewTree()
	assert.Nil(t, tr.ReplaceOrInsert(kv("a", "b")))
	assert.Nil(t, tr.isBST())
	assert.Nil(t, tr.ReplaceOrInsert(kv("c", "d")))
	assert.Nil(t, tr.isBST())
	assert.Nil(t, tr.ReplaceOrInsert(kv("e", "f")))
	assert.Nil(t, tr.isBST())
	assert.Equal(t, tr.Get(kv("c", "")).(keyValue), kv("c", "d"))
	assert.Equal(t, tr.Get(kv("a", "")).(keyValue), kv("a", "b"))
	assert.NotNil(t, tr.Delete(kv("a", "")))
	assert.Nil(t, tr.Delete(kv("a", "")))
	assert.NotNil(t, tr.Delete(kv("c", "")))
	assert.Nil(t, tr.ReplaceOrInsert(kv("a", "b")))
	assert.NotNil(t, tr.ReplaceOrInsert(kv("a", "d")))
	for i := 0; i < 100; i++ {
		assert.Nil(t, tr.ReplaceOrInsert(kv(strconv.Itoa(i), "foo")))
	}
	for i := 0; i < 100; i++ {
		if tr.Delete(kv(strconv.Itoa(i), "")) == nil {
			t.Fatalf("Failed to delete %v %v", i, tr)
		}
	}
}

// // func TestRandom(t *testing.T) {
// // 	const N = 4096
// // 	m := make(map[float64]float64)
// // 	tr := NewTree(CompareFloats)

// // 	seed := time.Now().UnixNano()
// // 	t.Logf("seed: %v", seed)
// // 	//rand.Seed(seed)
// // 	for i := 0; i < N; i++ {
// // 		k, v := rand.Float64(), rand.Float64()
// // 		m[k] = v
// // 		if tr.ReplaceOrInsert(k, v) {
// // 			t.Fatalf("%v", tr)
// // 		}
// // 	}
// // 	verified := 0
// // 	verify := func() {

// // 		for k, v := range m {
// // 			var it Iterator
// // 			if !it.Seek(tr, k) {
// // 				t.Fatalf("verify %v: Failed to find %v", verified, k)
// // 			}
// // 			if it.Value() != v {
// // 				t.Fatalf("Unexpected value for %v %v != %v", k, v, it.Value())
// // 			}
// // 		}
// // 		verified++
// // 	}

// // 	for i := 0; i < N; i++ {
// // 		// Use for loop for pseudorandom access.
// // 		var k, v float64
// // 		for k, v = range m {
// // 			break
// // 		}
// // 		var it Iterator
// // 		if !it.Seek(tr, k) {
// // 			t.Fatalf("did not find %v", k)
// // 		}
// // 		assert.Equal(t, it.Value(), v)
// // 		r := rand.Float64()
// // 		switch {
// // 		case r < .1:
// // 			assert.True(t, tr.Delete(k))
// // 			delete(m, k)
// // 			verify()
// // 		case r < .2:
// // 			k, v = rand.Float64(), rand.Float64()
// // 			assert.Nil(t, tr.ReplaceOrInsert(k, v))
// // 			m[k] = v
// // 		default:
// // 			v = rand.Float64()
// // 			m[k] = v
// // 			assert.True(t, tr.ReplaceOrInsert(k, v))
// // 		}
// // 	}
// // }
