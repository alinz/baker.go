package docker

type Event struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
}
