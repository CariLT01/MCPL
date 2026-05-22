package exp

import (
	"encoding/base64"
	"fmt"
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/data"
	"os"

	"crypto/ed25519"

	"github.com/golang-jwt/jwt/v5"
)

func CheckToken() {
	pubByes, err := base64.StdEncoding.DecodeString(config.PUBLIC_KEY)
	if err != nil {
		fmt.Println("Cannot decode public key")
		os.Exit(1)
	}
	pubKey := ed25519.PublicKey(pubByes)
	token, err := jwt.Parse(data.ISSUED_TOKEN, func(t *jwt.Token) (interface{}, error) {
		// Ensure algorithm is EdDSA
		if t.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		fmt.Println("Invalid token: ", err)
		os.Exit(1)
	}
}
