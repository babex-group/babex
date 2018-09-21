package babex

import (
	"encoding/json"
)

type CatchData struct {
	Error    string          `json:"error"`
	Exchange string          `json:"exchange"`
	Key      string          `json:"key"`
	Data     json.RawMessage `json:"data"`
}
