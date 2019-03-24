package rankorder

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func (t *Tree) isBST() error {
	return t.root.isBST(t, nil, nil)
}

func TestTree(t *testing.T) {
	tr := NewTree(func(a, b interface{}) int {
		return strings.Compare(a.(string), b.(string))
	})

	assert.False(t, tr.Upsert("a", "b"))
	assert.Nil(t, tr.isBST())
	assert.False(t, tr.Upsert("c", "d"))
	assert.Nil(t, tr.isBST())
	assert.False(t, tr.Upsert("e", "f"))
	assert.Nil(t, tr.isBST())
	var it Iterator
	assert.True(t, it.Seek(tr, "c"))
	assert.Equal(t, it.Value(), "d")
	assert.True(t, it.Seek(tr, "a"))
	assert.Equal(t, it.Value(), "b")
	assert.True(t, tr.Delete("a"))
	assert.False(t, tr.Delete("a"))
	assert.True(t, tr.Delete("c"))
	assert.False(t, tr.Upsert("a", "b"))
	assert.True(t, tr.Upsert("a", "d"))
	for i := 0; i < 100; i++ {
		assert.False(t, tr.Upsert(strconv.Itoa(i), "foo"))
	}
	for i := 0; i < 100; i++ {
		if !tr.Delete(strconv.Itoa(i)) {

			t.Fatalf("Failed to delete %v %v", i, tr)
		}
	}
}

func graphIt(tr *Tree) {
	plt, err := tr.Plot()
	if err != nil {
		panic(err)
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	f.Write(plt)
	fmt.Println(f.Name())
}

func TestRandom(t *testing.T) {
	const N = 4096
	m := make(map[float64]float64)
	tr := NewTree(CompareFloats)

	seed := time.Now().UnixNano()
	t.Logf("seed: %v", seed)
	//rand.Seed(seed)
	for i := 0; i < N; i++ {
		k, v := rand.Float64(), rand.Float64()
		m[k] = v
		if tr.Upsert(k, v) {
			t.Fatalf("%v", tr)
		}
	}
	verified := 0
	verify := func() {

		for k, v := range m {
			var it Iterator
			if !it.Seek(tr, k) {
				t.Fatalf("verify %v: Failed to find %v", verified, k)
			}
			if it.Value() != v {
				t.Fatalf("Unexpected value for %v %v != %v", k, v, it.Value())
			}
		}
		verified++
	}

	for i := 0; i < N; i++ {
		// Use for loop for pseudorandom access.
		var k, v float64
		for k, v = range m {
			break
		}
		var it Iterator
		if !it.Seek(tr, k) {
			t.Fatalf("did not find %v", k)
		}
		assert.Equal(t, it.Value(), v)
		r := rand.Float64()
		switch {
		case r < .1:
			assert.True(t, tr.Delete(k))
			delete(m, k)
			verify()
		case r < .2:
			k, v = rand.Float64(), rand.Float64()
			assert.False(t, tr.Upsert(k, v))
			m[k] = v
		default:
			v = rand.Float64()
			m[k] = v
			assert.True(t, tr.Upsert(k, v))
		}
	}
}
