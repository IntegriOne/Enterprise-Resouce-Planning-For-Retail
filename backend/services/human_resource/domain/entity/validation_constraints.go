package entity

import "time"

// Logika tiga nilai menjaga semantik NULL CHECK PostgreSQL pada kode domain murni.
type ruleTruth uint8

const (
	ruleUnknown ruleTruth = iota
	ruleFalse
	ruleTrue
)

func ruleKnown(value bool) ruleTruth {
	if value {
		return ruleTrue
	}
	return ruleFalse
}
func rulePresent(present bool, predicate func() bool) ruleTruth {
	if !present {
		return ruleUnknown
	}
	return ruleKnown(predicate())
}
func ruleAnd(a, b ruleTruth) ruleTruth {
	if a == ruleFalse || b == ruleFalse {
		return ruleFalse
	}
	if a == ruleUnknown || b == ruleUnknown {
		return ruleUnknown
	}
	return ruleTrue
}
func ruleOr(a, b ruleTruth) ruleTruth {
	if a == ruleTrue || b == ruleTrue {
		return ruleTrue
	}
	if a == ruleUnknown || b == ruleUnknown {
		return ruleUnknown
	}
	return ruleFalse
}
func ruleNot(a ruleTruth) ruleTruth {
	if a == ruleUnknown {
		return a
	}
	return ruleKnown(a == ruleFalse)
}
func ruleTruthCompare(a, b ruleTruth, op string) ruleTruth {
	if a == ruleUnknown || b == ruleUnknown {
		return ruleUnknown
	}
	if op == "=" {
		return ruleKnown(a == b)
	}
	return ruleKnown(a != b)
}
func ruleCompareResult(result int, op string) bool {
	switch op {
	case "=":
		return result == 0
	case "<>":
		return result != 0
	case ">":
		return result > 0
	case ">=":
		return result >= 0
	case "<":
		return result < 0
	case "<=":
		return result <= 0
	}
	return false
}

// CompareDates membandingkan tanggal kalender tanpa durasi 24 jam atau pengaruh DST.
func CompareDates(a, b time.Time) int {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	av := ay*10000 + int(am)*100 + ad
	bv := by*10000 + int(bm)*100 + bd
	if av < bv {
		return -1
	}
	if av > bv {
		return 1
	}
	return 0
}
