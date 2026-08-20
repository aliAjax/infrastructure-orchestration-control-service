package adapter

type NodeDTO struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

type GraphDTO struct {
	Nodes  []NodeDTO  `json:"nodes"`
	Levels [][]string `json:"levels"`
}
