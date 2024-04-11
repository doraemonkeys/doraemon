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
func TestSimpleKV_Set2(t *testing.T) {
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

	t.Run("Set key-value pair", func(t *testing.T) {
		key := "testKey"
		value := "testValue"

		err := db.Set(key, value)
		if err != nil {
			t.Errorf("Failed to set value: %v", err)
		}

		result, exists := db.Get(key)
		if !exists {
			t.Errorf("Key should exist after setting it")
		}
		if result != value {
			t.Errorf("Get returned incorrect value: got %v, want %v", result, value)
		}
	})

	t.Run("Set key-value pair with existing key", func(t *testing.T) {
		key := "testKey"
		value := "testValue2"

		err := db.Set(key, value)
		if err != nil {
			t.Errorf("Failed to set value: %v", err)
		}

		result, exists := db.Get(key)
		if !exists {
			t.Errorf("Key should exist after setting it")
		}
		if result != value {
			t.Errorf("Get returned incorrect value: got %v, want %v", result, value)
		}
	})

	t.Run("Set key-value pair with error", func(t *testing.T) {
		key := "testKey"
		value := "testValue3"

		// Simulate an error during saving
		f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		err = f.Chmod(0)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		err = db.Set(key, value)
		if err == nil {
			t.Error("Expected an error but got nil")
		}

		_, exists := db.Get(key)
		if exists {
			t.Error("Key should not exist after error")
		}
	})
}
