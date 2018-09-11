package babex

type CatchData struct {
	Error    error  `json:"error"`
	Exchange string `json:"exchange"`
	Key      string `json:"key"`
}
