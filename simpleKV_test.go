package doraemon

import (
	"os"
	"testing"
)

func TestNewSimpleKV(t *testing.T) {

	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{"1", "./test.db", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSimpleKV(tt.dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSimpleKV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSimpleKV_Set(t *testing.T) {
	fileName := "./test.db"
	err := os.WriteFile(fileName, []byte{}, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileName)
	db, err := NewSimpleKV(fileName)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := db.Get("test")
	if ok {
		t.Fatal("Get() error")
	}
	err = db.Set("test", "test")
	if err != nil {
		t.Fatal(err)
	}
	value, ok := db.Get("test")
	if !ok {
		t.Fatal("Get() error")
	}
	if value != "test" {
		t.Fatal("Set() error")
	}
	err = db.Set("test2", "test2")
	if err != nil {
		t.Fatal(err)
	}
	value, ok = db.Get("test2")
	if !ok {
		t.Fatal("Get() error")
	}
	if value != "test2" {
		t.Fatal("Set() error")
	}
	err = db.Delete("test")
	if err != nil {
		t.Fatal(err)
	}
	_, ok = db.Get("test")
	if ok {
		t.Fatal("Get() error")
	}
}

func TestSimpleKV(t *testing.T) {
	fileName := "./test.db"
	err := os.WriteFile(fileName, []byte{}, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileName)

	kv, err := NewSimpleKV(fileName)
	if err != nil {
		t.Fatalf("Failed to create SimpleKV: %v", err)
	}

	testKey := "testKey"
	testValue := "testValue"
	err = kv.Set(testKey, testValue)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	value, exists := kv.Get(testKey)
	if !exists || value != testValue {
		t.Errorf("Get returned incorrect value: got %v, want %v", value, testValue)
	}

	err = kv.Delete(testKey)
	if err != nil {
		t.Errorf("Failed to delete key: %v", err)
	}

	_, exists = kv.Get(testKey)
	if exists {
		t.Errorf("Key should have been deleted")
	}

	testKey2 := "testKey2"
	testValue2 := "testValue2"
	err = kv.Set(testKey2, testValue2)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	kv2, err := NewSimpleKV(fileName)
	if err != nil {
		t.Fatalf("Failed to create SimpleKV from existing db: %v", err)
	}

	_, exists = kv2.Get(testKey)
	if exists {
		t.Errorf("Key should not exist after restart")
	}

	value, exists = kv2.Get(testKey2)
	if !exists || value != testValue2 {
		t.Errorf("Get returned incorrect value: got %v, want %v", value, testValue2)
	}
}
