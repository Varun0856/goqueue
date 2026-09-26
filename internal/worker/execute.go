package worker

import (
	"errors"
	"slices"
	"strings"
)

func Execute(jobType string, payload string) (string, error) {
	switch jobType {
	case "uppercase":
		modified := strings.ToUpper(payload)
		return modified, nil

	case "reverse":
		reversed := Reverse(payload)
		return reversed, nil
	default:
		return "", errors.New("unknown job type")
	}
}

func Reverse(payload string) string {
	runes := []rune(payload)
	slices.Reverse(runes)
	return string(runes)
}
