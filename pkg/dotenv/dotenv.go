package dotenv

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

var (
	printer = func(str string) {
		log.Println(str)
	}
)

const (
	envFileName         = ".env"
	overrideEnvFileName = ".env.override"
)

// Load envs from envFileName
// If env already exists it will not be overwritten
func Load() {
	if err := readFile(envFileName, false); err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
		default:
			printer(fmt.Sprintf("loading envs: %s", err))
		}
	}
}

// Overload the same as Load, but additionally
// overrides ENVs from `.env.override` file
func Overload() {
	Load()

	if err := readFile(overrideEnvFileName, true); err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
		default:
			printer(fmt.Sprintf("overriding envs: %s", err))
		}
	}
}

func SetPrinter(p func(str string)) {
	printer = p
}

func searchCallerFile() string {
	_, file, _, _ := runtime.Caller(1)
	currentFile := file
	for i := 2; file == currentFile; i++ {
		_, file, _, _ = runtime.Caller(i)
	}

	return file
}

func readFile(name string, override bool) error {
	dir := findAppTomlDir(searchCallerFile())
	if dir == "" {
		return errors.New("can't find app.toml in project dir")
	}

	envFile := filepath.Join(dir, name)
	_, err := os.Stat(envFile)
	if err != nil {
		return err
	}

	load := godotenv.Load
	if override {
		load = godotenv.Overload
		envs, _ := godotenv.Read(envFile)
		for env := range envs {
			printer(fmt.Sprintf("Env %s will be overwritten", env))
		}
	}

	if err := load(envFile); err != nil {
		return fmt.Errorf("loading %s file, %s", envFile, err)
	}

	return nil
}

func findAppTomlDir(from string) string {
	dir := filepath.Dir(from)
	gopath := filepath.Clean(os.Getenv("GOPATH"))
	for dir != "/" && dir != gopath {
		appTomlFile := filepath.Join(dir, "app.toml")
		if _, err := os.Stat(appTomlFile); os.IsNotExist(err) {
			dir = filepath.Dir(dir)
			continue
		}
		return dir
	}
	return ""
}
