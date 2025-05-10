package tokens

type token struct {
	Access  string  `json:"token"`
	Refresh string  `json:"refresh"`
	Expires expires `json:"expires"`
}

type expires struct {
	Access  int `json:"access"`
	Refresh int `json:"refresh"`
}
