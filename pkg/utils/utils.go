package utils

import (
	"fmt"
	"os/user"
	"path/filepath"
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

// ExpandTilde expands the ~ in the file path to the home directory
func ExpandTilde(path string) (string, error) {
	if rest, ok := strings.CutPrefix(path, "~"); ok {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, rest), nil
	}
	return path, nil
}
