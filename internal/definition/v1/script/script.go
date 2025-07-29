package script

type Execs struct {
	Shells []Shell `json:"shells"`
}

type Shell struct {
	Name string `json:"name"`
}
