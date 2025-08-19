package fixpacks

type integrityResult struct {
	FixpackMd5  string `json:"fixpackMd5"`
	ExpectedMd5 string `json:"expectedMd5"`
}
