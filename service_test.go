package babex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentItem(t *testing.T) {
	index, item := getCurrentItem([]*ChainItem{
		{
			Key:        "one",
			Exchange:   "all",
			Successful: false,
		},
	})

	assert.Equal(t, 0, index)
	assert.NotNil(t, item)
	assert.Equal(t, "one", item.Key)
}

func TestGetCurrentItem1(t *testing.T) {
	index, item := getCurrentItem([]*ChainItem{
		{
			Key:        "one",
			Exchange:   "all",
			Successful: true,
		},
	})

	assert.Equal(t, -1, index)
	assert.Nil(t, item)
}