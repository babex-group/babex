package babex

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApplyWhen_Meta(t *testing.T) {
	meta := Meta{"one": "1"}
	when := When{"$meta.one": "1"}

	res, err := ApplyWhen(when, &InitialMessage{Meta: meta})

	assert.Nil(t, err)
	assert.Equal(t, true, res)

	meta["one"] = "2"

	res, err = ApplyWhen(when, &InitialMessage{Meta: meta})

	assert.Nil(t, err)
	assert.Equal(t, false, res)
}
