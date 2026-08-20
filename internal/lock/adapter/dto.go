package adapter

type LockDTO struct {
	Key     string `json:"key"`
	Owner   string `json:"owner"`
	Expires string `json:"expires"`
}
