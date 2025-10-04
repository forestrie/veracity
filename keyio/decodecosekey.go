package keyio

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"

	"github.com/veraison/go-cose"
)

// Cose Key as defined in: https://www.rfc-editor.org/rfc/rfc8152.html#page-33
//
//	COSE_Key = {
//		1 => tstr / int,          ; kty
//		? 2 => bstr,              ; kid
//		? 3 => tstr / int,        ; alg
//		? 4 => [+ (tstr / nt) ], ; key_ops
//		? 5 => bstr,              ; Base IV
//		* label => values
//	}
const (

	// TODO: these are private in go-cose
	KeyTypeLabel       = 1
	KeyIDLabel         = 2
	AlgorithmLabel     = 3
	KeyOperationsLabel = 4

	// :w
	// ECCurveLabel = -1
	// :w
	// ECXLabel     = -2
	// :w
	// ECYLabel     = -3
	// :w
	// ECDLabel     = -4

)

var (
	ErrKeyFormatError = errors.New("key format error")
)

type DecodedPublic struct {
	Alg    cose.Algorithm
	Public *ecdsa.PublicKey
}
type DecodedPrivate struct {
	Alg     cose.Algorithm
	Private *ecdsa.PrivateKey
}

func COSEDecodeEC2Private(
	m map[int64]any,
) (DecodedPrivate, error) {
	decoded, err := COSEDecodeEC2Public(m)
	if err != nil {
		return DecodedPrivate{}, fmt.Errorf("%w: decoding public component of private key.", err)
	}

	privateKey := &ecdsa.PrivateKey{
		PublicKey: *decoded.Public,
		D:         big.NewInt(0),
	}
	privateKey.D.SetBytes(m[cose.KeyLabelEC2D].([]byte))

	return DecodedPrivate{Alg: decoded.Alg, Private: privateKey}, nil
}

func COSEDecodeEC2Public(
	m map[int64]any,
) (DecodedPublic, error) {

	if err := KTYRequireEC2(m[KeyTypeLabel]); err != nil {
		return DecodedPublic{}, fmt.Errorf("failed to decode public key from map: %w", err)
	}

	alg, err := AlgRequireECDSA(m[AlgorithmLabel])
	if err != nil {
		return DecodedPublic{}, fmt.Errorf("failed to decode public key from map: %w", err)
	}

	// Key Type, must be EC2 or "EC". The string "EC" is accepted as an accommodation for JOSE

	x, err := DecodeLabeledBytes(m, cose.KeyLabelEC2X)
	if err != nil {
		return DecodedPublic{}, fmt.Errorf("failed to decode x coordinate from map: %w", err)
	}

	y, err := DecodeLabeledBytes(m, cose.KeyLabelEC2Y)
	if err != nil {
		return DecodedPublic{}, fmt.Errorf("failed to decode y coordinate from map: %w", err)
	}

	curve, err := COSEDecodeEC2Curve(m)
	if err != nil {
		return DecodedPublic{}, err
	}

	// TODO: As extras KeyIDLabel, KeyIDOps

	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     big.NewInt(0),
		Y:     big.NewInt(0),
	}
	publicKey.X.SetBytes(x)
	publicKey.Y.SetBytes(y)

	return DecodedPublic{Alg: alg, Public: &publicKey}, nil
}

// COSEDecodeEC2Curve require the curve to be appropriate for EC2 type keys
// And return a representation appropriate for building a golang EC public key
func COSEDecodeEC2Curve(m map[int64]any) (elliptic.Curve, error) {

	curveAny, ok := m[cose.KeyLabelEC2Curve]
	if !ok {
		return nil, fmt.Errorf("missing curve label in COSE key map")
	}

	curve, ok := curveAny.(int64)
	if !ok {
		ucurve, ok := curveAny.(uint64)
		if !ok {
			return nil, fmt.Errorf("wrong type for curve label in COSE key map, needed int64, got %T", curveAny)
		}
		curve = int64(ucurve)
	}
	switch cose.Curve(curve) {
	case cose.CurveP256:
		return elliptic.P256(), nil
	case cose.CurveP384:
		return elliptic.P384(), nil
	case cose.CurveP521:
		return elliptic.P521(), nil
	default:
		return nil, fmt.Errorf("unsupported curve label in COSE key map: %d", curve)
	}
}

// KTYRequireEC2 returns an error if the label is not EC2
// The strings "EC" and "EC2" are accepted as an accommodation for JOSE
// Both uint64 and int64 are accepted as accommodations for sloppy encoders.
// Per https://www.rfc-editor.org/rfc/rfc8152.html#section-13
func KTYRequireEC2(label any) error {
	s, ok := label.(string)

	if ok {
		if s != "EC" && s != "EC2" {
			return fmt.Errorf("%w: expected EC or EC2 or %d, got %s", ErrKeyFormatError, cose.KeyTypeEC2, s)
		}
		return nil
	}

	i64, ok := label.(int64)
	if !ok {
		u64, ok := label.(uint64)
		if !ok {
			return fmt.Errorf("%w: expected [uint64|int64|string] not %T", ErrKeyFormatError, label)
		}
		i64 = int64(u64)
	}
	if cose.KeyType(i64) != cose.KeyTypeEC2 {
		return fmt.Errorf("%w: expected EC or EC2 or %d, got %d", ErrKeyFormatError, cose.KeyTypeEC2, i64)
	}
	return nil
}

func AlgRequireECDSA(label any) (cose.Algorithm, error) {
	s, ok := label.(string)
	if ok {
		switch s {
		case "ES256":
			return cose.Algorithm(cose.AlgorithmES256), nil
		case "ES384":
			return cose.Algorithm(cose.AlgorithmES384), nil
		case "ES512":
			return cose.Algorithm(cose.AlgorithmES512), nil
		default:
			return 0, fmt.Errorf("%w: decoding string label, expected ES256, ES384 or ES512, got %s", ErrKeyFormatError, s)
		}
	}

	i64, ok := label.(int64)
	if !ok {
		u64, ok := label.(uint64)
		if !ok {
			return 0, fmt.Errorf("%w: decoding integer label expected [uint64|int64] not %T", ErrKeyFormatError, label)
		}
		i64 = int64(u64)
	}

	switch cose.Algorithm(i64) {
	case cose.AlgorithmES256:
		return cose.Algorithm(i64), nil
	case cose.AlgorithmES384:
		return cose.Algorithm(i64), nil
	case cose.AlgorithmES512:
		return cose.Algorithm(i64), nil
	default:
		return 0, fmt.Errorf(
			"%w: decoding integer label expected %d, %d or %d, got %d",
			ErrKeyFormatError, cose.AlgorithmES256, cose.AlgorithmES384, cose.AlgorithmES512, i64)
	}
}

//
// label decode helper
//

func DecodeLabeledBytes(m map[int64]interface{}, label int64) ([]byte, error) {
	v, ok := m[label]
	if !ok {
		return nil, fmt.Errorf("missing label %d in map", label)
	}
	return DecodeBytes(v)
}

func DecodeBytes(label any) ([]byte, error) {
	b, ok := label.([]byte)
	if ok {
		return b, nil
	}

	s, ok := label.(string)
	if ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("%w: expected []byte or string, got %T", ErrKeyFormatError, label)
}
