package doraemon

import (
	"encoding/json"
	"os"
	"sync"
)

// var Config = NewConfigWrapper[string]("")

type ConfigWrapper[T comparable] struct {
	ConfigMu   sync.RWMutex
	Config     T
	filePath   string
	jsonIndent string
}

func NewConfigWrapper2[T comparable](filePath string) (*ConfigWrapper[T], error) {
	var v T
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &ConfigWrapper[T]{
		filePath: filePath,
		Config:   v,
	}, nil
}

func NewConfigWrapper[T comparable](filePath string) *ConfigWrapper[T] {
	c, err := NewConfigWrapper2[T](filePath)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *ConfigWrapper[T]) Path() string {
	return c.filePath
}

func (c *ConfigWrapper[T]) SetJsonIndent(jsonIndent string) {
	c.jsonIndent = jsonIndent
}

func (c *ConfigWrapper[T]) Save() error {
	return c.SaveTo(c.filePath)
}

func (c *ConfigWrapper[T]) SaveTo(filePath string) error {
	if c.jsonIndent == "" {
		c.jsonIndent = "    "
	}

	c.ConfigMu.RLock()
	defer c.ConfigMu.RUnlock()

	jsonData, err := json.MarshalIndent(c.Config, "", c.jsonIndent)
	if err != nil {
		return err
	}

	return WriteFilePreservePerms(filePath, jsonData)
}

// Reload will cause panic, Do not use this function.
//
// If you need to reload the configuration, please create a new ConfigWrapper.
func (c *ConfigWrapper[T]) Reload() error {
	panic("don't use this function")
}
