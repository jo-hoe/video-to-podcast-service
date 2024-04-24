package common

import "reflect"

func ValueOrDefault[T any](value, defaultValue T) T {
	reflectedValue := reflect.ValueOf(value)
	if reflectedValue.Kind() == reflect.Invalid {
		return defaultValue
	} else if reflectedValue.Kind() == reflect.String && reflectedValue.String() == "" {
		return defaultValue
	}
	return value
}