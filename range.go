package icinga

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type (
	// Range is a combination of a lower boundary, an upper boundary
	// and a flag for inverted (@) range semantics. See [0] for more
	// details.
	Range interface {
		Check(float64) bool
		CheckInt(int) bool
		CheckInt32(int32) bool
	}

	rangeImpl struct {
		Start  float64
		End    float64
		Invert bool
	}
)

// NewRange parse a string and returns a Range object
// 10			< 0 or > 10, (outside the range of {0 .. 10})
// 10:		< 10, (outside {10 .. ∞})
// ~:10		> 10, (outside the range of {-∞ .. 10})
// 10:20	< 10 or > 20, (outside the range of {10 .. 20})
// @10:20	≥ 10 and ≤ 20, (inside the range of {10 .. 20})
func NewRange(value string) (Range, error) {
	// Set defaults
	r := &rangeImpl{
		Start:  0,
		End:    math.Inf(1),
		Invert: false,
	}

	value = strings.Trim(value, " \n\r")

	// We can override a default value with an empty string and use 0 as range
	if len(value) == 0 {
		return r, nil
	}

	// Check for inverted semantics
	if value[0] == '@' {
		r.Invert = true
		value = value[1:]
	}

	// Parse lower limit
	endPos := strings.Index(value, ":")
	if endPos > -1 {
		if value[0] == '~' {
			r.Start = math.Inf(-1)
		} else {
			min, err := strconv.ParseFloat(value[0:endPos], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse lower limit: %v", err)
			}
			r.Start = min
		}
		value = value[endPos+1:]
	}

	// Parse upper limit
	if len(value) > 0 {
		max, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse upper limit: %v", err)
		}
		r.End = max
	}

	if r.End < r.Start {
		return nil, errors.New("Invalid range definition. min <= max violated")
	}

	// OK
	return r, nil
}

// Check returns true if an alert should be raised based on the range (if the
// value is outside the range for normal semantics, or if the value is
// inside the range for inverted semantics ('@-semantics')).
func (r *rangeImpl) Check(value float64) bool {
	// Ranges are treated as a closed interval.
	if r.Start <= value && value <= r.End {
		return r.Invert
	}
	return !r.Invert
}

// CheckInt is a convenience method which does an unchecked type
// conversion from an int to a float64.
func (r *rangeImpl) CheckInt(val int) bool {
	return r.Check(float64(val))
}

// CheckInt32 is a convenience method which does an unchecked type
// conversion from an int32 to a float64.
func (r *rangeImpl) CheckInt32(val int32) bool {
	return r.Check(float64(val))
}

func (r *rangeImpl) CheckValue(val interface{}) bool {
	return r.Check(val.(float64))
}
