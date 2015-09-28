package to

import "time"

// MayBool converts v to bool or returns val.
func MayBool(v interface{}, val bool) bool {
	b, err := Bool(v)
	if err != nil {
		return val
	}
	return b
}

// MayDuration converts v to Duration or returns val.
func MayDuration(v interface{}, val time.Duration) time.Duration {
	d, err := Duration(v)
	if err != nil {
		return val
	}
	return d
}

// MayInt converts v to int64 or returns val.
func MayInt(v interface{}, val int) int {
	i, err := Int64(v)
	if err != nil {
		return val
	}
	return int(i)
}

// MayInt64 converts v to int64 or returns val.
func MayInt64(v interface{}, val int64) int64 {
	i, err := Int64(v)
	if err != nil {
		return val
	}
	return i
}

// MayFloat converts v to float64 or returns val.
func MayFloat(v interface{}, val float64) float64 {
	f, err := Float64(v)
	if err != nil {
		return val
	}
	return f
}

// MayMap converts v to map[string]interface{} or returns val.
func MayMap(v interface{}, val map[string]interface{}) map[string]interface{} {
	m, err := Map(v)
	if err != nil {
		return val
	}
	return m
}

// MaySlice converts v to []interface{} or returns val.
func MaySlice(v interface{}, val []interface{}) []interface{} {
	sli, err := Slice(v)
	if err != nil {
		return val
	}
	return sli
}

// MayString converts v to string or returns ""
func MayString(v interface{}) string {
	return String(v)
}

// MayTime converts v to Time or returns Time{}
func MayTime(v interface{}, val time.Time) time.Time {
	t, err := Time(v)
	if err != nil {
		return val
	}
	return t
}
