package rankorder

import (
	"encoding/base64"
	"fmt"
	"math/rand"
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
}

func TestRandom(t *testing.T) {
	const N = 4096
	m := make(map[float64]float64)
	tr := NewTree(func(a, b interface{}) (cmp int) {
		if af, bf := a.(float64), b.(float64); af < bf {
			cmp = -1
		} else if bf < af {
			cmp = 1
		}
		return cmp
	})
	defer func() {
		if r := recover(); r != nil {
			plt, err := tr.Plot()
			if err != nil {
				panic(err)
			}
			dst := make([]byte, base64.StdEncoding.EncodedLen(len(plt)))
			base64.StdEncoding.Encode(dst, plt)
			fmt.Println(string(dst))
			panic(r)
		}

	}()
	seed := time.Now().UnixNano()
	t.Logf("seed: %v", seed)
	//rand.Seed(seed)
	for i := 0; i < N; i++ {
		k, v := rand.Float64(), rand.Float64()
		m[k] = v
		assert.False(t, tr.Upsert(k, v))
	}

	for i := 0; i < N; i++ {
		// Use for loop for pseudorandom access.
		var k, v float64
		for k, v = range m {
			break
		}
		var it Iterator
		assert.True(t, it.Seek(tr, k))
		assert.Equal(t, it.Value(), v)
		r := rand.Float64()
		switch {
		case r < .1:
			assert.True(t, tr.Delete(k))
		case r < .2:
			k, v = rand.Float64(), rand.Float64()
			m[k] = v
		default:
			v = rand.Float64()
			m[k] = v
			assert.True(t, tr.Upsert(k, v))
		}
	}
}
