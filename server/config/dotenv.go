package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"

	"github.com/joho/godotenv"
)

// dotEnvPaths is ordered by precedence, relative to the server/ working directory.
var dotEnvPaths = []string{".env", "../.env"}

// LoadDotEnv loads optional env files. godotenv never overrides a variable that
// is already set, so real environment variables win, then earlier files.
func LoadDotEnv() error {
	loaded := false
	for _, path := range dotEnvPaths {
		err := godotenv.Load(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		log.Printf("Loaded env file %s", path)
		loaded = true
	}
	if !loaded {
		log.Println("No .env file found, using system environment")
	}
	return nil
}
