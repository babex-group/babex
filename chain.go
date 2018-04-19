package babex

type ChainItem struct {
	Successful bool   `json:"successful"`
	Exchange   string `json:"exchange"`
	Key        string `json:"key"`
	IsMultiple bool   `json:"isMultiple"`
}
