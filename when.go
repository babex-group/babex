package babex

import (
	"errors"
	"strings"
)

const (
	WhenEntityMeta = "$meta"
)

var (
	ErrorWhenInvalidKey   = errors.New("when: invalid key")
	ErrorWhenInvalidValue = errors.New("when: invalid value")
)

type When map[string]interface{}

func ApplyWhen(when When, meta Meta) (bool, error) {
	for k, w := range when {
		keyParts := strings.Split(k, ".")

		if len(keyParts) == 0 {
			return true, ErrorWhenInvalidKey
		}

		from := keyParts[0]

		var expectedValues []string
		var value string

		switch v := w.(type) {
		case string:
			expectedValues = []string{v}
		case []string:
			expectedValues = v
		}

		switch from {
		case WhenEntityMeta:
			if len(keyParts) < 2 {
				return true, ErrorWhenInvalidValue
			}

			metaKey := keyParts[1]
			value = meta[metaKey]
		}

		var isValid bool

		for _, expected := range expectedValues {
			if value == expected {
				isValid = true
				break
			}
		}

		if !isValid {
			return false, nil
		}
	}

	return true, nil
}
