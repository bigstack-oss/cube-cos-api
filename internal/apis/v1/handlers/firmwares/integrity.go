package firmwares

type integrityResult struct {
	FirmwareMd5 string `json:"firmwareMd5"`
	ExpectedMd5 string `json:"expectedMd5"`
}
