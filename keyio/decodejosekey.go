package keyio

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/veraison/go-cose"
)

// JWK represents a single JOSE key (simplified for EC public keys)
type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Alg string `json:"alg,omitempty"`
	Use string `json:"use,omitempty"`
	Kid string `json:"kid,omitempty"`
}

// JWKS represents a JOSE key set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// ReadECDSAPublicJOSE decodes a JSON-encoded JOSE EC public key set and returns the first *last* key as DecodedPublic
func ReadECDSAPublicJOSE(fileName string) (DecodedPublic, error) {

	joseKey, err := os.ReadFile(fileName)
	if err != nil {
		return DecodedPublic{}, fmt.Errorf("failed to read public keyset file: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(joseKey, &jwks); err != nil {
		return DecodedPublic{}, err
	}
	if len(jwks.Keys) == 0 {
		return DecodedPublic{}, errors.New("no keys found in JWKS")
	}
	jwk := jwks.Keys[len(jwks.Keys)-1]
	if jwk.Kty != "EC" {
		return DecodedPublic{}, errors.New("only EC keys are supported")
	}
	// Decode base64url-encoded X and Y
	// Use base64.RawURLEncoding to decode JOSE base64url values (no padding)
	x, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return DecodedPublic{}, err
	}
	y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return DecodedPublic{}, err
	}

	var curv elliptic.Curve
	switch jwk.Crv {
	case "P-256":
		curv = elliptic.P256()
	case "P-384":
		curv = elliptic.P384()
	case "P-521":
		curv = elliptic.P521()
	default:
		return DecodedPublic{}, fmt.Errorf("%w: curve %s invalid for EC keys", ErrKeyFormatError, jwk.Crv)
	}

	var alg cose.Algorithm
	switch jwk.Alg {
	case "PS256":
		alg = cose.AlgorithmPS256
	case "PS384":
		alg = cose.AlgorithmPS384
	case "PS512":
		alg = cose.AlgorithmPS512
	case "ES256":
		alg = cose.AlgorithmES256
	case "ES384":
		alg = cose.AlgorithmES384
	case "ES512":
		alg = cose.AlgorithmES512
	default:
		return DecodedPublic{}, fmt.Errorf("%w: alg %s invalid for EC keys", ErrKeyFormatError, jwk.Alg)
	}

	publicKey := ecdsa.PublicKey{
		Curve: curv,
		X:     big.NewInt(0),
		Y:     big.NewInt(0),
	}
	publicKey.X.SetBytes(x)
	publicKey.Y.SetBytes(y)

	decoded := DecodedPublic{
		Public: &publicKey,
		Alg:    alg,
	}
	return decoded, nil
}
