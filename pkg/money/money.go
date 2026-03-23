package money

import (
	"fmt"
	"math"
)

const precision = 100

type Money int64

func FromFloat(f float64) Money {
	return Money(math.Round(f * precision))
}

func FromInt(i int) Money {
	return Money(i * precision)
}

func (m Money) Float() float64 {
	return float64(m) / precision
}

func (m Money) Int() int {
	return int(m / precision)
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f", m.Float())
}

func (m Money) Add(other Money) Money {
	return m + other
}

func (m Money) Sub(other Money) Money {
	return m - other
}

func (m Money) Mul(multiplier float64) Money {
	return Money(math.Round(float64(m) * multiplier))
}

func (m Money) IsPositive() bool {
	return m > 0
}

func (m Money) IsNegative() bool {
	return m < 0
}

func (m Money) IsZero() bool {
	return m == 0
}

func AlmostEqual(a, b Money, tolerance Money) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
