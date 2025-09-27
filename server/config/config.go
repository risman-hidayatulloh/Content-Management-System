package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

var AppConfig Config

func Load() Config {
	// 1) ENV override (recommended in container)
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return mustLoadFrom(p)
	}

	// 2) CWD (./config.toml) – cocok utk binary di /app
	if _, err := os.Stat("./config.toml"); err == nil {
		return mustLoadFrom("./config.toml")
	}

	// 3) Fallback: relatif ke file source (cara lama)
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to get the current filename")
	}
	filePath := filepath.Join(filename, "../..")
	configPath := fmt.Sprintf("%s/%s", path.Join(path.Dir(filePath)), "config.toml")
	return mustLoadFrom(configPath)
}

func mustLoadFrom(configPath string) Config {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var z Config
	if err := toml.Unmarshal(b, &z); err != nil {
		log.Fatal(err)
	}
	AppConfig = z
	return z
}
