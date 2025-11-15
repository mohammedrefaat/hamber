package utils

import (
	"errors"
	"strconv"
)

// ParseUint converts a string to uint
func ParseUint(s string) (uint, error) {
	if s == "" {
		return 0, errors.New("empty string cannot be converted to uint")
	}

	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(val), nil
}

// ParseInt converts a string to int
func ParseInt(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string cannot be converted to int")
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// ParseInt64 converts a string to int64
func ParseInt64(s string) (int64, error) {
	if s == "" {
		return 0, errors.New("empty string cannot be converted to int64")
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// MustParseUint converts a string to uint, panics on error
func MustParseUint(s string) uint {
	val, err := ParseUint(s)
	if err != nil {
		panic(err)
	}
	return val
}
