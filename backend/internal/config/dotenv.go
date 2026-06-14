package config

import (
	"bufio"
	"os"
	"strings"
)

// loadDotEnv reads a .env file (if present in the working directory) and sets
// any variables that are not already defined in the real environment. This is a
// convenience for local development; in production (Cloud Run) the variables
// come from the environment directly and no .env file exists.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // no .env file: nothing to do
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Strip optional surrounding quotes.
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
