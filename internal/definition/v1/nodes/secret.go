package nodes

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func GenToken(hostname string) string {
	hash := sha512.Sum512([]byte(hostname + auths.DefaultOidcClientSecret))
	return hex.EncodeToString(hash[:])
}

func GetSecretHeaders() map[string]string {
	return map[string]string{
		"Node":          base.Hostname,
		"Authorization": fmt.Sprintf("Bearer %s", auths.DefaultNodeToken),
	}
}
