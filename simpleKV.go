package doraemon

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

type SimpleKV struct {
	data     map[string]string
	dataLock sync.RWMutex
	dbPath   string
}

func NewSimpleKV(dbPath string) (*SimpleKV, error) {
	f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	kv := &SimpleKV{
		data:     make(map[string]string),
		dataLock: sync.RWMutex{},
		dbPath:   dbPath,
	}
	err = json.NewDecoder(f).Decode(&kv.data)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return kv, nil
}

func (kv *SimpleKV) save() error {
	jsonData, err := json.MarshalIndent(kv.data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(kv.dbPath, jsonData, 0644)
}

func (kv *SimpleKV) Get(key string) (string, bool) {
	kv.dataLock.RLock()
	defer kv.dataLock.RUnlock()
	value, ok := kv.data[key]
	return value, ok
}

func (kv *SimpleKV) Delete(key string) error {
	kv.dataLock.Lock()
	defer kv.dataLock.Unlock()
	v, ok := kv.data[key]
	if !ok {
		return nil
	}
	delete(kv.data, key)
	if err := kv.save(); err != nil {
		kv.data[key] = v
		return err
	}
	return nil
}

func (kv *SimpleKV) Set(key, value string) error {
	kv.dataLock.Lock()
	defer kv.dataLock.Unlock()
	kv.data[key] = value
	if err := kv.save(); err != nil {
		delete(kv.data, key)
		return err
	}
	return nil
}

func (kv *SimpleKV) SetIfNotExist(key, value string) (bool, error) {
	kv.dataLock.Lock()
	defer kv.dataLock.Unlock()
	if _, ok := kv.data[key]; ok {
		return false, nil
	}
	kv.data[key] = value
	if err := kv.save(); err != nil {
		delete(kv.data, key)
		return false, err
	}
	return true, nil
}

func (kv *SimpleKV) Range(f func(key, value string) bool) {
	kv.dataLock.RLock()
	defer kv.dataLock.RUnlock()
	for k, v := range kv.data {
		if !f(k, v) {
			break
		}
	}
}
