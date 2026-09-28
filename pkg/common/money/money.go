package money

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Amount represents monetary value in minor units (e.g., Paise or Cents) to avoid float inaccuracies.
// ₹100.50 is represented as 10050.
type Amount int64

// FromDecimal converts a standard decimal float (e.g., 100.50) into minor units (10050).
func FromDecimal(val float64) Amount {
	return Amount(math.Round(val * 100))
}

// FromString parses a string representation of decimal currency (e.g. "100.50" or "100") into minor units.
func FromString(s string) (Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid monetary amount format: %w", err)
	}
	return FromDecimal(f), nil
}

// ToDecimal converts minor units back to a float64 representation.
func (a Amount) ToDecimal() float64 {
	return float64(a) / 100.0
}

// String formats the amount as a decimal currency string (e.g., "100.50").
func (a Amount) String() string {
	sign := ""
	val := a
	if val < 0 {
		sign = "-"
		val = -val
	}
	integerPart := val / 100
	fractionalPart := val % 100
	return fmt.Sprintf("%s%d.%02d", sign, integerPart, fractionalPart)
}

// GORM & SQL Scanner/Valuer implementations for Amount
func (a Amount) Value() (driver.Value, error) {
	return int64(a), nil
}

func (a *Amount) Scan(value any) error {
	if value == nil {
		*a = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*a = Amount(v)
	case int32:
		*a = Amount(v)
	case int:
		*a = Amount(v)
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*a = Amount(i)
	default:
		return fmt.Errorf("cannot scan %T into Amount", value)
	}
	return nil
}
