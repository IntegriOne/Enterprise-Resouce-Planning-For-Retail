package entity

import (
	"math/big"
	"strings"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// Rat mengembalikan salinan rasional eksak; pemanggil tidak dapat mengubah Decimal asal.
func (d Decimal) Rat() (*big.Rat, error) {
	if len(d) > 128 || !decimalPattern.MatchString(string(d)) {
		return nil, hrerrors.Invalid("decimal", "must be a finite decimal")
	}
	number, ok := new(big.Rat).SetString(string(d))
	if !ok {
		return nil, hrerrors.Invalid("decimal", "must be a finite decimal")
	}
	return number, nil
}

// ValidatePrecision menolak nilai yang akan dibulatkan atau overflow oleh numeric(p,s).
func (d Decimal) ValidatePrecision(precision, scale int) error {
	if precision < 1 || scale < 0 || scale > precision {
		return hrerrors.Invalid("decimal", "invalid precision configuration")
	}
	if len(d) > 128 || !decimalPattern.MatchString(string(d)) {
		return hrerrors.Invalid("decimal", "must be a bounded finite decimal")
	}
	parts := strings.SplitN(strings.TrimLeft(string(d), "+-"), ".", 2)
	integerDigits := len(strings.TrimLeft(parts[0], "0"))
	fractionalDigits := 0
	if len(parts) == 2 {
		fractionalDigits = len(strings.TrimRight(parts[1], "0"))
	}
	if integerDigits > precision-scale || fractionalDigits > scale {
		return hrerrors.Invalid("decimal", "exceeds numeric precision or scale")
	}
	return nil
}

// DecimalSum menjumlahkan nilai pada skala eksplisit tanpa pembulatan diam-diam.
func DecimalSum(scale int, values ...Decimal) (Decimal, error) {
	if scale < 0 || scale > 8 {
		return "", hrerrors.Invalid("decimal", "unsupported scale")
	}
	total := new(big.Rat)
	for _, value := range values {
		if err := value.ValidatePrecision(38, scale); err != nil {
			return "", err
		}
		n, err := value.Rat()
		if err != nil {
			return "", err
		}
		total.Add(total, n)
	}
	return Decimal(total.FloatString(scale)), nil
}

// DecimalNegate membentuk nilai lawan yang eksak.
func DecimalNegate(value Decimal) (Decimal, error) {
	n, err := value.Rat()
	if err != nil {
		return "", err
	}
	if n.Sign() == 0 {
		return "0", nil
	}
	if strings.HasPrefix(string(value), "-") {
		return Decimal(strings.TrimPrefix(string(value), "-")), nil
	}
	return Decimal("-" + strings.TrimPrefix(string(value), "+")), nil
}

func ruleNumber(value string) *big.Rat {
	n, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil
	}
	return n
}
func ruleAdd(a, b *big.Rat) *big.Rat {
	if a == nil || b == nil {
		return nil
	}
	return new(big.Rat).Add(a, b)
}
func ruleSubtract(a, b *big.Rat) *big.Rat {
	if a == nil || b == nil {
		return nil
	}
	return new(big.Rat).Sub(a, b)
}
func ruleNumberCompare(a, b *big.Rat, op string) bool {
	if a == nil || b == nil {
		return false
	}
	return ruleCompareResult(a.Cmp(b), op)
}
