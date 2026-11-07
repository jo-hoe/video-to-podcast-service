package common

func ValueOrDefault[T any](value, defaultValue T) T {
	switch v := any(value).(type) {
	case string:
		if v == "" {
			return defaultValue
		}
		return value
	default:
		if any(value) == nil {
			return defaultValue
		}
		return value
	}
}
