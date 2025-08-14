package firmwares

type integrityResult struct {
	FirmwareMd5 string `json:"firmareMd5"`
	ExpectedMd5 string `json:"expectedMd5"`
}
