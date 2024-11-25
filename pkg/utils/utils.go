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
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir := usr.HomeDir
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}
