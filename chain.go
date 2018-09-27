package babex

type ChainItem struct {
	Successful bool              `json:"successful"`
	Exchange   string            `json:"exchange"`
	Key        string            `json:"key"`
	IsMultiple bool              `json:"isMultiple"`
	SetCount   Chain             `json:"setCount"`
	Catch      Chain             `json:"catch"`
	Meta       map[string]string `json:"meta"` // Meta for particular item
}

type Chain []ChainItem

func SetCurrentItemSuccess(chain Chain) Chain {
	newChain := make([]ChainItem, len(chain))

	copy(newChain, chain)

	index := getCurrentChainIndex(chain)
	newChain[index].Successful = true

	return newChain
}

func getCurrentChainIndex(chain Chain) int {
	for i, item := range chain {
		if item.Successful != true {
			return i
		}
	}

	return -1
}
