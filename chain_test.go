package babex

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetNextChainElement(t *testing.T) {
	chain := Chain{
		{
			Exchange: "first_topic",
		},
	}

	newChain := SetCurrentItemSuccess(chain)

	assert.Equal(t, true, newChain[0].Successful)
	assert.Equal(t, "first_topic", newChain[0].Exchange)
	assert.Equal(t, false, chain[0].Successful)
}
