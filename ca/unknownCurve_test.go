package ca

import (
	"crypto/elliptic"
	"math/big"
)

type (
	// unknownCurve is designed to make x509.oidFromNamedCurve fail.
	unknownCurve struct{}
)

func (u unknownCurve) Params() *elliptic.CurveParams {
	return nil
}

func (u unknownCurve) IsOnCurve(x, y *big.Int) bool {
	return false
}

func (u unknownCurve) Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(0)
}

func (u unknownCurve) Double(x1, y1 *big.Int) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(0)
}

func (u unknownCurve) ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(0)
}

func (u unknownCurve) ScalarBaseMult(k []byte) (x, y *big.Int) {
	return big.NewInt(0), big.NewInt(0)
}
