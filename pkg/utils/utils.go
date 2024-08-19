package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var hashRegex = regexp.MustCompile("^[a-fA-F0-9]{40}$")

func ValidateHash(hashes []string) error {
	var invalid []string

	for _, hash := range hashes {
		if !hashRegex.MatchString(hash) {
			invalid = append(invalid, hash)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("invalid hashes: %s", strings.Join(invalid, ","))
	}

	return nil
}
