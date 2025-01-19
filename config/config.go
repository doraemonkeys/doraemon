package config

import (
	"os"
	"reflect"

	"github.com/doraemonkeys/doraemon"
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
