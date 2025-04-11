package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"reflect"
	"testing"
)

func TestAES(t *testing.T) {

	var newer []func(key string) (SymmetricCipher, error)
	newer = append(newer, func(key string) (SymmetricCipher, error) {
		return NewAESGCMFromHex(key)
	})
	newer = append(newer, func(key string) (SymmetricCipher, error) {
		return NewAESCBCFromHex(key)
	})
	newer = append(newer, func(key string) (SymmetricCipher, error) {
		return NewAESCBCFromHex2(key)
	})

	type args struct {
		plainText string
	}
	// 256 bits
	key := "A0A1A2A3A4A5A6A7A8A9AAABACADAEAF"
	for _, newCipher := range newer {
		Crypter, err := newCipher(key)
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
}

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

func TestNewAESGCM(t *testing.T) {
	key := "thisis32bitkey1234thisis32bitkey" // 32 bytes = 256 bits
	plaintext := "This is a secret message!"

	// Encryption
	aesgcm, err := NewAESGCM([]byte(key))
	if err != nil {
		t.Error("Error creating AESGCM:", err)
		return
	}

	ciphertext, err := aesgcm.Encrypt([]byte(plaintext))
	if err != nil {
		t.Error("Encryption error:", err)
		return
	}
	t.Logf("Ciphertext: %x\n", ciphertext)
	ciphertextCopy := make([]byte, len(ciphertext))
	copy(ciphertextCopy, ciphertext)
	// Decryption
	decrypted, err := aesgcm.Decrypt(ciphertext)
	if err != nil {
		t.Error("Decryption error:", err)
		return
	}
	t.Logf("Decrypted: %s\n", decrypted)
	if bytes.Equal(ciphertext, ciphertextCopy) {
		t.Error("expect ciphertext to be modified")
	}

	// Example using Hex key
	hexKey := "74686973697333326269746b65793132333474686973697333326269746b6579" // Hex representation of the key above
	aesgcmHex, err := NewAESGCMFromHex(hexKey)
	if err != nil {
		t.Error("Error creating AESGCM from hex:", err)
		return
	}
	ciphertextHex, err := aesgcmHex.Encrypt([]byte(plaintext))
	if err != nil {
		t.Error("Encryption error (hex key):", err)
		return
	}
	t.Logf("Ciphertext (hex key): %x\n", ciphertextHex)

	decryptedHex, err := aesgcmHex.Decrypt(ciphertextHex)
	if err != nil {
		t.Error("Decryption error (hex key):", err)
		return
	}
	t.Logf("Decrypted (hex key): %s\n", decryptedHex)

	// Example with specific nonce

	nonce := make([]byte, aesgcm.AEAD.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Error("Error making nonce:", err)
		return
	}

	nonce = append(nonce, []byte(plaintext)...)
	nonceCopy := make([]byte, len(nonce))
	copy(nonceCopy, nonce)
	ciphertextWithNonce, err := aesgcm.EncryptWithNonce(nonce)
	if err != nil {
		t.Error("Encryption error with nonce:", err)
		return
	}
	t.Logf("Ciphertext with nonce: %x\n", ciphertextWithNonce)
	if !bytes.Equal(nonce, nonceCopy) {
		t.Error("plaintext should not be modified")
	}

	decryptedWithNonce, err := aesgcm.Decrypt(ciphertextWithNonce)
	if err != nil {
		t.Error("Decryption error with nonce:", err)
		return
	}
	t.Logf("Decrypted with nonce: %s\n", decryptedWithNonce)

	if string(decryptedWithNonce) != plaintext {
		t.Errorf("Decrypted with nonce: %s, want %s", decryptedWithNonce, plaintext)
	}
}

// TestNewAESGCM tests the NewAESGCM constructor.
func TestNewAESGCM2(t *testing.T) {
	key := make([]byte, 32) // Valid 256-bit key
	rand.Read(key)          // Fill with random data

	_, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("NewAESGCM failed: %v", err)
	}

	// Test with invalid key lengths.
	invalidKey1 := make([]byte, 16) // too short
	_, err = NewAESGCM(invalidKey1)
	if err != nil {
		t.Error("NewAESGCM failed with a 16-byte key")
	}
	invalidKey2 := make([]byte, 20) // invalid length
	_, err = NewAESGCM(invalidKey2)
	if err == nil {
		t.Error("NewAESGCM should have failed with a 20-byte key, but it didn't.")
	}

	invalidKey3 := make([]byte, 64) // too long
	_, err = NewAESGCM(invalidKey3)
	if err == nil {
		t.Error("NewAESGCM should have failed with a 64-byte key, but it didn't.")
	}
}

// TestNewAESGCMFromHex tests the NewAESGCMFromHex constructor.
func TestNewAESGCMFromHex(t *testing.T) {
	// Test case 1: Valid key
	validHexKey := "6368616e676520746869732070617373776f726420746f206120736563726574" // 32-byte (256-bit) key
	_, err := NewAESGCMFromHex(validHexKey)
	if err != nil {
		t.Errorf("NewAESGCMFromHex failed with a valid key: %v", err)
	}

	// Test case 2: Invalid hex string (non-hex characters)
	invalidHexKey1 := "xyz123"
	_, err = NewAESGCMFromHex(invalidHexKey1)
	if err == nil {
		t.Errorf("NewAESGCMFromHex should have failed with invalid hex characters")
	}

	//Test case 4: Invalid hex string (incorrect length, too long)
	invalidHexKey3 := "6368616e676520746869732070617373776f726420746f20612073656372657470617373776f726420746f206120736563726574"
	_, err = NewAESGCMFromHex(invalidHexKey3)
	if err == nil {
		t.Error("NewAESGCMFromHex should have failed with an invalid key length")
	}

}

// TestEncryptDecrypt tests the Encrypt and Decrypt methods.
func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	c, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("NewAESGCM failed: %v", err)
	}

	plaintext := []byte("this is a secret message")
	ciphertext, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := c.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted text does not match original plaintext. Got: %s, Expected: %s", decrypted, plaintext)
	}

	//Test case 2: Empty string
	plaintext2 := []byte("")
	ciphertext2, err := c.Encrypt(plaintext2)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	decrypted2, err := c.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !bytes.Equal(plaintext2, decrypted2) {
		t.Errorf("Decrypted text does not match original plaintext. Got: %s, Expected: %s", decrypted2, plaintext2)
	}

	// Test case 3: Decrypting with wrong key.
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)
	cWrong, _ := NewAESGCM(wrongKey)
	_, err = cWrong.Decrypt(ciphertext) // ciphertext was generated using c, not cWrong.
	if err == nil {
		t.Error("Decrypt should have failed with an incorrect key, but didn't.")
	}

	//Test case 4: Tampered ciphertext
	tamperedCiphertext := make([]byte, len(ciphertext))
	copy(tamperedCiphertext, ciphertext)
	tamperedCiphertext[len(tamperedCiphertext)-1] ^= 0xFF // Modify a byte.

	_, err = c.Decrypt(tamperedCiphertext)
	if err == nil {
		t.Error("Decrypt should have failed with tampered ciphertext, but didn't")
	}

	//Test case 5: short ciphertext
	shortCiphertext := []byte{1, 2, 3}
	_, err = c.Decrypt(shortCiphertext)
	if err == nil {
		t.Error("Decrypt should have failed with short ciphertext, but didn't")
	}

}

func TestEncryptWithNonce(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	c, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("NewAESGCM failed: %v", err)
	}

	nonceSize := c.NonceSize()
	data := make([]byte, nonceSize+len("this is a secret message"))
	copy(data[nonceSize:], []byte("this is a secret message"))
	rand.Read(data[:nonceSize]) // Fill the nonce part with random data

	ciphertext, err := c.EncryptWithNonce(data)
	if err != nil {
		t.Fatalf("EncryptWithNonce failed: %v", err)
	}

	decrypted, err := c.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(data[nonceSize:], decrypted) {
		t.Errorf("Decrypted text does not match original plaintext. Got: %s, Expected: %s", decrypted, data[nonceSize:])
	}

	//Test Case 2: Empty string
	data2 := make([]byte, nonceSize)
	rand.Read(data2)
	ciphertext2, err := c.EncryptWithNonce(data2)
	if err != nil {
		t.Fatalf("EncryptWithNonce failed: %v", err)
	}
	decrypted2, err := c.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal([]byte{}, decrypted2) {
		t.Errorf("Decrypted text does not match original plaintext. Got: %s, Expected: %s", decrypted2, []byte{})
	}

	//Test Case 3: Data is too short
	shortData := make([]byte, nonceSize-1)
	_, err = c.EncryptWithNonce(shortData)
	if err == nil {
		t.Error("EncryptWithNonce should have failed, but didn't")
	}

	//Test Case 4: Decrypt with wrong key
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)
	cWrong, _ := NewAESGCM(wrongKey)
	_, err = cWrong.Decrypt(ciphertext) // Decrypt using the wrong key
	if err == nil {
		t.Error("Decrypt should have failed with the wrong key, but didn't")
	}
}

func TestEncryptAuth(t *testing.T) {
	key, _ := hex.DecodeString("6368616e676520746869732070617373") // Example key
	c, err := NewAESGCM(key)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name           string
		data           []byte
		additionalData [][]byte
		expectedError  bool
	}{
		{
			name:           "Empty data, no additional data",
			data:           []byte{},
			additionalData: [][]byte{},
			expectedError:  false,
		},
		{
			name:           "Non-empty data, no additional data",
			data:           []byte("some data to encrypt"),
			additionalData: [][]byte{},
			expectedError:  false,
		},
		{
			name:           "Empty data, with additional data",
			data:           []byte{},
			additionalData: [][]byte{[]byte("auth data")},
			expectedError:  false,
		},
		{
			name:           "Non-empty data, with additional data",
			data:           []byte("some other data"),
			additionalData: [][]byte{[]byte("auth data 1"), []byte("auth data 2")},
			expectedError:  false,
		},
		{
			name: "Non-empty data, multiple additional data",
			data: []byte("test message"),
			additionalData: [][]byte{
				[]byte("additional data 1"),
				[]byte("additional data 2"),
				[]byte("additional data 3"),
			},
			expectedError: false,
		},
		{ // check if different key/nonce produce different ciphertext
			name:           "Non-empty data, check for consistent ciphertext",
			data:           []byte("consistent test data"),
			additionalData: [][]byte{[]byte("consistent auth data")},
			expectedError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := c.EncryptAuth(tc.data, tc.additionalData...)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error: %v, got: %v", tc.expectedError, err)
			}
			if tc.expectedError {
				return // If an error was expected, we're done.
			}

			if len(ciphertext) < c.AEAD.NonceSize() {
				t.Errorf("ciphertext is shorter than nonce size")
			}

			// Decrypt and check if it matches the original data
			plaintext, err := c.DecryptAuth(ciphertext, tc.additionalData...)
			if err != nil {
				t.Fatalf("DecryptAuth failed: %v", err)
			}
			if !bytes.Equal(plaintext, tc.data) {
				t.Errorf("decrypted data does not match original data.  Expected: %x, got: %x", tc.data, plaintext)
			}

			if tc.name == "Non-empty data, check for consistent ciphertext" {
				// Generate another cipher for same data, check if the output changes
				anotherCiphertext, anotherErr := c.EncryptAuth(tc.data, tc.additionalData...)
				if anotherErr != nil {
					t.Fatalf("EncryptAuth 2nd try failed: %v", anotherErr)
				}
				if bytes.Equal(ciphertext, anotherCiphertext) {
					t.Errorf("Repeated calls to encryptAuth are producing identical ciphertexts")
				}
			}
		})
	}
}

func TestEncryptAuth_DifferentKeys(t *testing.T) {
	key1, _ := hex.DecodeString("6368616e676520746869732070617373")
	key2, _ := hex.DecodeString("736968742065676e6168632073736170") // Different key
	c1, err := NewAESGCM(key1)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := NewAESGCM(key2)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("test data")
	additionalData := [][]byte{[]byte("auth data")}

	ciphertext1, err := c1.EncryptAuth(data, additionalData...)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext2, err := c2.EncryptAuth(data, additionalData...)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Ciphertexts with different keys are the same")
	}

	// Attempt to decrypt with the wrong key.
	_, err = c2.DecryptAuth(ciphertext1, additionalData...)
	if err == nil {
		t.Error("Decryption succeeded with the wrong key")
	}

	_, err = c1.DecryptAuth(ciphertext2, additionalData...)
	if err == nil {
		t.Error("Decryption succeeded with the wrong key")
	}

}

// Helper function to create a broken AEAD (for negative testing).
func createBrokenAEAD() (cipher.AEAD, error) {
	key := make([]byte, 32) // Use a valid key size for aes.NewCipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a GCM with a bad tag size
	gcm, err := cipher.NewGCMWithTagSize(block, 8) // Valid tag sizes are 12-16
	if err != nil {                                // Expect an error here
		return nil, err
	}

	return gcm, nil
}

func TestEncryptAuth_BrokenAEAD(t *testing.T) {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	c, err := NewAESGCM(key)
	if err != nil {
		t.Fatal(err)
	}

	brokenAEAD, err := createBrokenAEAD()
	if err == nil {
		//replace original aead with broken one
		c.AEAD = brokenAEAD

		_, err := c.EncryptAuth([]byte("test data"), [][]byte{[]byte("auth data")}...)
		if err == nil {
			t.Error("Expected error, but got none")
		}
	} //else the helper function already return err, skip

	var f = func() {
		defer func() {
			r := recover()
			if r == nil && err == nil {
				t.Error("Expected panic or error, but got none")
			}
		}()
		//force nonce error
		c, _ = NewAESGCM(key) //reset
		c.AEAD = &FaultyReaderAEAD{AEAD: c.AEAD}
		_, err = c.EncryptAuth([]byte("some data"))
		if err == nil {
			t.Error("Expected an error due to faulty nonce generation, but got nil")
		}
	}

	// Test that the function panics when the error is nil
	f()

}

// FaultyReaderAEAD is a wrapper around cipher.AEAD that simulates a failure in nonce generation.
type FaultyReaderAEAD struct {
	cipher.AEAD
}

func (f *FaultyReaderAEAD) NonceSize() int {
	return f.AEAD.NonceSize()
}

func (f *FaultyReaderAEAD) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	return f.AEAD.Open(dst, nonce, ciphertext, additionalData)
}

func (f *FaultyReaderAEAD) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	//simulate error
	return f.AEAD.Seal(dst, []byte("bad_nonce"), plaintext, additionalData)
}

// Helper for broken reader test
type brokenReader struct{}

func (br brokenReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestEncryptAuth_ReaderFailure(t *testing.T) {
	key := make([]byte, 32)
	c, err := NewAESGCM(key)
	if err != nil {
		t.Fatal(err)
	}

	// Replace the random reader with our broken reader.
	oldReader := rand.Reader
	rand.Reader = brokenReader{}
	defer func() { rand.Reader = oldReader }() // Restore the original reader.

	_, err = c.EncryptAuth([]byte("data"))
	if err == nil {
		t.Error("Expected error from reader failure, but got nil")
	}
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected ErrUnexpectedEOF, got %v", err)
	}
}
