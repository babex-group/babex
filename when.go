package babex

import (
	"errors"
	"strings"
)

const (
	CommandMeta = "$meta"
)

var (
	ErrorInvalidKey   = errors.New("invalid key")
	ErrorInvalidValue = errors.New("invalid value")
)

type When map[string]interface{}

func ApplyWhen(when When, msg *InitialMessage) (bool, error) {
	for k, w := range when {
		keyParts := strings.Split(k, ".")

		if len(keyParts) == 0 {
			return true, ErrorInvalidKey
		}

		switch keyParts[0] {
		case CommandMeta:
			if len(keyParts) < 2 {
				return true, ErrorInvalidKey
			}

			metaKey := keyParts[1]
			metaValue := msg.Meta[metaKey]
			expectedValue, ok := w.(string)
			if !ok {
				return true, ErrorInvalidValue
			}

			if metaValue != expectedValue {
				return false, nil
			}
		}
	}

	return true, nil
}
