package config

import (
	"os"
	"reflect"

	"github.com/doraemonkeys/doraemon"
	"github.com/joho/godotenv"
	toml "github.com/pelletier/go-toml/v2"
)

func InitJsonConfig[T any](configFile string, createDefault func(path string) error) (*T, error) {
	return doraemon.InitJsonConfig[T](configFile, createDefault)
}

func InitTomlConfig[T any](configFile string, createDefault func(path string) error) (*T, error) {
	var config T

	if createDefault == nil {
		createDefault = func(path string) error {
			c, err := toml.Marshal(doraemon.DeepCreateEmptyInstance(reflect.TypeOf(config)))
			if err != nil {
				return err
			}
			return os.WriteFile(path, c, 0666)
		}
	}
	return doraemon.InitConfig[T](configFile, createDefault, toml.Unmarshal)
}

// LoadEnv loads environment variables from the given files, if they exist.
// It does NOT override existing environment variables. Use .env for defaults.
// Environment variable names are case-sensitive.
func LoadEnv(path ...string) error {
	paths := []string{}
	for _, p := range path {
		if doraemon.FileIsExist(p).IsTrue() {
			paths = append(paths, p)
		}
	}
	return godotenv.Load(paths...)
}

// OverloadEnv loads environment variables from the given files, if they exist.
// It WILL override existing environment variables.  Use this to force specific values.
func OverloadEnv(path ...string) error {
	paths := []string{}
	for _, p := range path {
		if doraemon.FileIsExist(p).IsTrue() {
			paths = append(paths, p)
		}
	}
	return godotenv.Overload(paths...)
}
