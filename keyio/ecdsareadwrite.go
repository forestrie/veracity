package keyio

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/go-cose"
)

const (
	ECDSAPublicDefaultPEMFileName  = "ecdsa-key-public.pem"
	ECDSAPrivateDefaultPEMFileName = "ecdsa-key-private.pem"
	ECDSAPublicDefaultFileName     = "ecdsa-key-public.cbor"
	ECDSAPrivateDefaultFileName    = "ecdsa-key-private.cbor"
	ECDSAPrivateDefaultPerm        = 0600 // Default permission for private key file
	ECDSAPublicDefaultPerm         = 0644 // Default permission for private key file
)

func ReadECDSAPublicCOSE(
	fileName string,
) (DecodedPublic, error) {
	// Read the public key from the default file
	data, err := os.ReadFile(fileName)
	if err != nil {
		return DecodedPublic{}, fmt.Errorf("failed to read public key file: %w", err)
	}

	var m map[int64]interface{}
	if err := cbor.Unmarshal(data, &m); err != nil {
		return DecodedPublic{}, err
	}

	return COSEDecodeEC2Public(m)
}

func ReadECDSAPrivateCOSE(
	fileName string,
	expectedStandardCurve ...string,
) (DecodedPrivate, error) {
	// Read the private key from the default file
	data, err := os.ReadFile(fileName)
	if err != nil {
		return DecodedPrivate{}, fmt.Errorf("failed to read private key file: %w", err)
	}
	var m map[int64]interface{}
	if err := cbor.Unmarshal(data, &m); err != nil {
		return DecodedPrivate{}, err
	}

	return COSEDecodeEC2Private(m)
}

func ReadECDSAPrivatePEM(filePath string) (DecodedPrivate, error) {
	pemData, err := os.ReadFile(filePath)
	if err != nil {
		return DecodedPrivate{}, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return DecodedPrivate{}, errors.New("invalid PEM block or type")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return DecodedPrivate{}, err
	}

	coseKey, err := cose.NewKeyFromPrivate(key)
	if err != nil {
		return DecodedPrivate{}, err
	}
	decoded := DecodedPrivate{
		Private: key,
		Alg:     coseKey.Algorithm,
	}

	return decoded, nil
}

func ReadECDSAPublicPEM(filePath string) (DecodedPublic, error) {
	pemData, err := os.ReadFile(filePath)
	if err != nil {
		return DecodedPublic{}, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return DecodedPublic{}, errors.New("invalid PEM block or type")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return DecodedPublic{}, err
	}

	ecdsaKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return DecodedPublic{}, errors.New("not an ECDSA public key")
	}
	coseKey, err := cose.NewKeyFromPublic(ecdsaKey)
	if err != nil {
		return DecodedPublic{}, err
	}
	decoded := DecodedPublic{
		Public: ecdsaKey,
		Alg:    coseKey.Algorithm,
	}

	return decoded, nil
}

// Serializes the key to PEM format
func encodeECDSAPrivateKeyToPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(block), nil
}

// Writes PEM to a file with 0600 permissions
func WriteECDSAPrivatePEM(pemFile string, key *ecdsa.PrivateKey) error {
	pemBytes, err := encodeECDSAPrivateKeyToPEM(key)
	if err != nil {
		return fmt.Errorf("PEM encoding failed: %w", err)
	}
	return os.WriteFile(pemFile, pemBytes, 0600)
}

func WriteECDSAPublicCOSE(
	pubFile string,
	publicKey *ecdsa.PublicKey,
) (string, error) {
	var err error

	if _, err = WriteCoseECDSAPublicKey(pubFile, publicKey); err != nil {
		return "", err
	}
	return pubFile, nil
}

func WriteECDSAPrivateCOSE(
	privFile string,
	privateKey *ecdsa.PrivateKey,
) (string, error) {
	var err error

	if _, err = WriteCoseECDSAPrivateKey(privFile, privateKey); err != nil {
		return "", err
	}
	return privFile, nil
}

// Encode private key to COSE_Key format (as CBOR bytes)
func encodePrivateKeyToCOSE(key *ecdsa.PrivateKey) ([]byte, error) {
	m := map[int64]interface{}{
		int64(1):              int64(cose.KeyTypeEC2),
		int64(3):              cose.AlgorithmPS256,
		cose.KeyLabelEC2Curve: cose.CurveP256, // P-256
		cose.KeyLabelEC2X:     key.PublicKey.X.Bytes(),
		cose.KeyLabelEC2Y:     key.PublicKey.Y.Bytes(),
		cose.KeyLabelEC2D:     key.D.Bytes(),
	}
	return cbor.Marshal(m)
}

// Encode public key to COSE_Key format (as CBOR bytes)
func encodePublicKeyToCOSE(key *ecdsa.PublicKey) ([]byte, error) {
	m := map[int64]interface{}{
		int64(1):              int64(cose.KeyTypeEC2),
		int64(3):              cose.AlgorithmPS256,
		cose.KeyLabelEC2Curve: cose.CurveP256, // P-256
		cose.KeyLabelEC2X:     key.X.Bytes(),
		cose.KeyLabelEC2Y:     key.Y.Bytes(),
	}
	return cbor.Marshal(m)
}

func WriteCoseECDSAPrivateKey(
	fileName string,
	privateKey *ecdsa.PrivateKey,
	perms ...os.FileMode,
) ([]byte, error) {
	var err error
	var data []byte
	if data, err = encodePrivateKeyToCOSE(privateKey); err != nil {
		return nil, err
	}

	perm := os.FileMode(ECDSAPrivateDefaultPerm) // Default permission
	if len(perms) > 0 {
		perm = perms[0]
	}

	// Save to file
	if err := os.WriteFile(fileName, data, perm); err != nil {
		return nil, err
	}
	return data, nil
}

func WriteCoseECDSAPublicKey(
	fileName string,
	publicKey *ecdsa.PublicKey,
	perms ...os.FileMode,
) ([]byte, error) {
	var err error
	var data []byte
	if data, err = encodePublicKeyToCOSE(publicKey); err != nil {
		return nil, err
	}

	perm := os.FileMode(ECDSAPublicDefaultPerm) // Default permission
	if len(perms) > 0 {
		perm = perms[0]
	}

	// Save to file
	if err := os.WriteFile(fileName, data, perm); err != nil {
		return nil, err
	}
	return data, nil
}
