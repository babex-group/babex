package babex

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMeta_Merge(t *testing.T) {
	first := Meta{"a": "1"}
	second := Meta{"a": "2"}
	third := Meta{"b": "3"}

	first.Merge(second, third)

	assert.Equal(t, "2", first["a"])
	assert.Equal(t, "3", first["b"])
}
