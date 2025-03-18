package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"io"
	"reflect"
	"testing"
)

func TestAESCBC(t *testing.T) {
	type args struct {
		plainText string
	}
	// 256 bits
	key := "A0A1A2A3A4A5A6A7A8A9AAABACADAEAF"
	Crypter, err := NewAESCBCFromHex(key)
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"123456"}, false},
		{"", args{"1234567890123456789012345678901234567890123456789012345678901234567890"}, false},
		{"", args{"h"}, false},
		{"", args{"hello world"}, false},
		{"", args{"ghhhhhhhhhhhhhhg"}, false},
		{"", args{"你好，世界"}, false},
		{"", args{"hello world, 你好，世界"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Crypter.Encrypt([]byte(tt.args.plainText))
			if (err != nil) != tt.wantErr {
				t.Errorf("CbcAESCrypt.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			t.Logf("CbcAESCrypt.Encrypt() = %s", got)
			t.Logf("CbcAESCrypt.Encrypt() = %x", got)

			// 解密
			got2, err := Crypter.Decrypt(got)
			if err != nil {
				t.Errorf("CbcAESCrypt.Decrypt() error = %v", err)
				return
			}
			t.Logf("CbcAESCrypt.Decrypt() = %s", got2)

			if string(got2) != tt.args.plainText {
				t.Errorf("CbcAESCrypt.Decrypt() got2 = %v, want %v", got2, tt.args.plainText)
			}
		})
	}

	// 加大原文长度
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				tt.args.plainText += tt.args.plainText
			}
			got, err := Crypter.Encrypt([]byte(tt.args.plainText))
			if err != nil {
				if !tt.wantErr {
					t.Errorf("CbcAESCrypt.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// 解密
			got2, err := Crypter.Decrypt(got)
			if err != nil {
				t.Errorf("CbcAESCrypt.Decrypt() error = %v", err)
				return
			}

			if string(got2) != tt.args.plainText {
				t.Errorf("CbcAESCrypt.Decrypt() got2 = %v, want %v", got2, tt.args.plainText)
			}
		})
	}
}

// MockRandReader provides a deterministic reader for testing purposes.
type MockRandReader struct {
	Data []byte
	Pos  int
}

func (m *MockRandReader) Read(p []byte) (n int, err error) {
	if m.Pos >= len(m.Data) {
		return 0, io.EOF
	}
	n = copy(p, m.Data[m.Pos:])
	m.Pos += n
	return n, nil
}

func TestNewAESCBC(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "Valid Key (16 bytes)",
			key:     []byte("0123456789abcdef"),
			wantErr: false,
		},
		{
			name:    "Valid Key (24 bytes)",
			key:     []byte("0123456789abcdef01234567"),
			wantErr: false,
		},
		{
			name:    "Valid Key (32 bytes)",
			key:     []byte("0123456789abcdef0123456789abcdef"),
			wantErr: false,
		},
		{
			name:    "Invalid Key (Short)",
			key:     []byte("0123456789abcde"), // 15 bytes
			wantErr: true,
		},
		{
			name:    "Invalid Key (Long)",
			key:     []byte("0123456789abcdef0123456789abcdefg"), //33
			wantErr: true,
		},
		{
			name:    "Empty Key",
			key:     []byte{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAESCBC(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAESCBC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.key == nil {
				t.Errorf("NewAESCBC() = %v, want non-nil", got)
			}
			if !tt.wantErr { // Further checks if no error is expected
				if !reflect.DeepEqual(got.key, tt.key) {
					t.Errorf("NewAESCBC() key = %v, want %v", got.key, tt.key)
				}
				if got.Block == nil {
					t.Error("NewAESCBC().Block should not be nil")
				}
			}
		})
	}
}

func TestNewAESCBCFromHex(t *testing.T) {
	tests := []struct {
		name    string
		hexKey  string
		wantKey []byte
		wantErr bool
	}{
		{
			name:    "Valid Hex Key (16 bytes)",
			hexKey:  "00112233445566778899aabbccddeeff",
			wantKey: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			wantErr: false,
		},
		{
			name:    "Invalid Hex Key (Invalid Characters)",
			hexKey:  "00112233445566778899aabbccddeefg", // 'g' is invalid
			wantKey: nil,
			wantErr: true,
		},
		{
			name:    "Invalid Hex Key (Incorrect Length)",
			hexKey:  "00112233445566778899aabbccddeef", // one too short
			wantKey: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAESCBCFromHex(tt.hexKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAESCBCFromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.key == nil {
				t.Errorf("NewAESCBCFromHex() = %v, want non-nil", got)
			}
			if !tt.wantErr && !reflect.DeepEqual(got.key, tt.wantKey) {
				t.Errorf("NewAESCBCFromHex() key = %v, want %v", got.key, tt.wantKey)
			}
		})
	}
}

func TestAESCBC_EncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef") // 16-byte key
	c, err := NewAESCBC(key)
	if err != nil {
		t.Fatal(err)
	}
	// Replace the global rand.Reader with our mock reader for deterministic tests.
	originalRandReader := rand.Reader
	defer func() { rand.Reader = originalRandReader }()
	mockIV := []byte("abcdef0123456789") // Mock IV for testing
	rand.Reader = &MockRandReader{Data: mockIV}

	tests := []struct {
		name    string
		data    []byte
		want    []byte // Expected *plaintext* after decryption
		wantErr bool
	}{
		{
			name:    "Empty Data",
			data:    []byte{},
			want:    []byte{}, // Empty data should remain empty after padding and unpadding
			wantErr: false,
		},
		{
			name:    "Single Block Data",
			data:    []byte("short message"),
			want:    []byte("short message"),
			wantErr: false,
		},
		{
			name:    "Multi-Block Data",
			data:    []byte("this is a longer message that spans multiple blocks"),
			want:    []byte("this is a longer message that spans multiple blocks"),
			wantErr: false,
		},
		{
			name:    "Data is exactly one block long",
			data:    []byte("0123456789abcdef"),
			want:    []byte("0123456789abcdef"),
			wantErr: false,
		},
		{
			name:    "Data length is a multiple of block size",
			data:    []byte("0123456789abcdef0123456789abcdef"), // 32 bytes
			want:    []byte("0123456789abcdef0123456789abcdef"), // 32 bytes
			wantErr: false,
		},
		{
			name:    "Chinese characters",
			data:    []byte("你好世界"),
			want:    []byte("你好世界"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ciphertext, err := c.Encrypt(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCBC.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ciphertext == nil {
				t.Errorf("AESCBC.Encrypt() = %v, want non-nil", ciphertext)
			}

			if !tt.wantErr {
				if len(ciphertext) != len(PKCS7Padding(tt.data, c.BlockSize()))+c.BlockSize() {
					t.Errorf("AESCBC.Encrypt() length = %v, want %v", len(ciphertext), len(PKCS7Padding(tt.data, c.BlockSize()))+c.BlockSize())
				}
				//Check that the IV is prepended
				if !bytes.Equal(ciphertext[:c.BlockSize()], mockIV) {
					t.Errorf("AESCBC.Encrypt() IV = %v, want %v", ciphertext[:c.BlockSize()], mockIV)
				}
			}

			rand.Reader = &MockRandReader{Data: mockIV}

			plaintext, err := c.Decrypt(ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCBC.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && plaintext == nil {
				t.Errorf("AESCBC.Decrypt() = %v, want non-nil", plaintext)
			}
			if !tt.wantErr && !reflect.DeepEqual(plaintext, tt.want) {
				t.Errorf("AESCBC.Decrypt() = %v, want %v", plaintext, tt.want)
			}

		})
	}
}

func TestAESCBC_Decrypt_Error(t *testing.T) {
	key := []byte("0123456789abcdef") // 16-byte key
	c, err := NewAESCBC(key)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    bool
	}{
		{
			name:       "Invalid Ciphertext Length (Too Short)",
			ciphertext: []byte("short"), // Shorter than block size
			wantErr:    true,
		},
		{
			name:       "Invalid Padding",
			ciphertext: append([]byte("0123456789abcdef"), []byte("invalid padding")...), //block size + some bytes.
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCBC.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPKCS7PaddingUnpadding(t *testing.T) {
	blockSize := aes.BlockSize // Use AES block size (16 bytes)
	tests := []struct {
		name        string
		input       []byte
		blockSize   int
		padded      []byte
		unpadded    []byte
		expectError bool
	}{
		{
			name:      "Empty input",
			input:     []byte{},
			blockSize: blockSize,
			padded:    bytes.Repeat([]byte{byte(blockSize)}, blockSize),
			unpadded:  []byte{},
		},
		{
			name:      "Input shorter than block size",
			input:     []byte("hello"),
			blockSize: blockSize,
			padded:    []byte("hello\v\v\v\v\v\v\v\v\v\v\v"), // 11 is \v
			unpadded:  []byte("hello"),
		},
		{
			name:      "Input equal to block size",
			input:     []byte("0123456789abcdef"),
			blockSize: blockSize,
			padded:    append([]byte("0123456789abcdef"), bytes.Repeat([]byte{byte(blockSize)}, blockSize)...),
			unpadded:  []byte("0123456789abcdef"),
		},
		{
			name:        "Invalid padding (padding > block size)",
			input:       append([]byte("0123456789abcdef"), bytes.Repeat([]byte{byte(blockSize + 1)}, blockSize)...), // craft an invalid one.
			blockSize:   blockSize,
			expectError: true,
		},
		{
			name:        "Invalid padding zero",
			input:       append([]byte("0123456789abcdef"), bytes.Repeat([]byte{byte(0)}, blockSize)...),
			blockSize:   blockSize,
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := PKCS7Padding(tt.input, tt.blockSize)
			if len(tt.padded) > 0 && !reflect.DeepEqual(padded, tt.padded) {
				t.Errorf("PKCS7Padding() = %v, want %v", padded, tt.padded)
			}

			if !tt.expectError { //Only run unpadding if no error expected.
				unpadded, err := PKCS7UnPadding(padded, tt.blockSize)
				if err != nil {
					t.Fatalf("PKCS7UnPadding() unexpected error %v", err)
				}

				if !reflect.DeepEqual(unpadded, tt.unpadded) {
					t.Errorf("PKCS7UnPadding() = %v, want %v", unpadded, tt.unpadded)
				}
			}

			if tt.expectError {
				_, err := PKCS7UnPadding(tt.input, tt.blockSize) // Use input directly, which *should* have invalid padding.
				if err == nil {
					t.Error("PKCS7Unpadding() expected error, got nil")
				}
			}

		})
	}
}
