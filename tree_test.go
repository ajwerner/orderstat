package rankorder

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (t *Tree) isBST() error {
	return t.root.isBST(t, nil, nil)
}

type keyValue struct {
	k string
	v string
}

func (kv keyValue) Less(other interface{}) bool {
	return kv.k < other.(keyValue).k
}

func kv(k, v string) keyValue { return keyValue{k: k, v: v} }

type intItem int

func (ii intItem) Less(other interface{}) bool {
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
		// assert.Equal(t, tr.Len(), i+1)
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
	tr.AscendGreatorOrEqual(intItem(items[seen]), expect)
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
}

func (it *Iterator) isBST(t *Tree, min, max Item) error {
	if it.node == nil {
		return nil
	}
	if min != nil && it.k.Less(min) {
		return fmt.Errorf("key %v < min %v", it.k, min)
	}
	if max != nil && max.Less(it.k) {
		return fmt.Errorf("key %v > max %v", it.k, max)
	}
	l := it.l(t)
	if l.node != nil && it.k.Less(l.k) {
		return fmt.Errorf("parent key %v < left child key %v", it.k, l.k)
	}
	r := it.r(t)
	if r.node != nil && r.k.Less(it.k) {
		return fmt.Errorf("parent key (%v) %v > right child key (%v)", it.np, it.k, r.np)
	}
	if err := l.isBST(t, min, it.k); err != nil {
		return err
	}
	if err := r.isBST(t, it.k, max); err != nil {
		return err
	}
	return nil
}

func TestTree(t *testing.T) {
	tr := NewTree()
	assert.Nil(t, tr.ReplaceOrInsert(kv("a", "b")))
	assert.Nil(t, tr.isBST())
	assert.Nil(t, tr.ReplaceOrInsert(kv("c", "d")))
	assert.Nil(t, tr.isBST())
	assert.Nil(t, tr.ReplaceOrInsert(kv("e", "f")))
	assert.Nil(t, tr.isBST())
	var it Iterator
	assert.True(t, it.Seek(tr, kv("c", "")))
	assert.Equal(t, it.Item().(keyValue), kv("c", "d"))
	assert.True(t, it.Seek(tr, kv("a", "")))
	assert.Equal(t, it.Item().(keyValue), kv("a", "b"))
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
