package doraemon

import (
	"reflect"
	"testing"
)

func TestCbcAESCrypt(t *testing.T) {
	type args struct {
		plainText string
	}
	// 256 bits
	key := "A0A1A2A3A4A5A6A7A8A9AAABACADAEAF"
	Crypter, err := NewAESCryptFromHex(key)
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
		{"", args{""}, true},
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

func TestRSA_EncryptOAEP(t *testing.T) {
	type args struct {
		plainText []byte
		label     []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{[]byte("hello world"), []byte("label")}},
		{"", args{[]byte("hello world, 你好，世界"), []byte("label")}},
	}
	privateKey, publicKey, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherText, err := RSA_EncryptOAEP(tt.args.plainText, publicKey, tt.args.label)
			if err != nil {
				t.Errorf("RSA_EncryptOAEP() error = %v", err)
				return
			}
			got, err := RSA_DecryptOAEP(cipherText, privateKey, tt.args.label)
			if err != nil {
				t.Errorf("RSA_DecryptOAEP() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.args.plainText) {
				t.Errorf("RSA_DecryptOAEP() = %v, want %v", got, tt.args.plainText)
			}
		})
	}
	// 加大原文长度
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				tt.args.plainText = append(tt.args.plainText, tt.args.plainText...)
			}
			newPlainText := tt.args.plainText
			cipherText, err := RSA_EncryptOAEP(newPlainText, publicKey, tt.args.label)
			if err != nil {
				t.Errorf("RSA_EncryptOAEP() error = %v", err)
				return
			}
			got, err := RSA_DecryptOAEP(cipherText, privateKey, tt.args.label)
			if err != nil {
				t.Errorf("RSA_DecryptOAEP() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, newPlainText) {
				t.Errorf("RSA_DecryptOAEP() = %v, want %v", got, tt.args.plainText)
			}
		})
	}
}

func TestRSA_EncryptPKCS1v15(t *testing.T) {
	type args struct {
		plainText []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{[]byte("hello world")}},
		{"", args{[]byte("hello world, 你好，世界")}},
		{"", args{[]byte("gopher !!!")}},
	}
	privateKey, publicKey, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherText, err := RSA_EncryptPKCS1v15(tt.args.plainText, publicKey)
			if err != nil {
				t.Errorf("RSA_EncryptPKCS1v15() error = %v", err)
				return
			}
			got, err := RSA_DecryptPKCS1v15(cipherText, privateKey)
			if err != nil {
				t.Errorf("RSA_DecryptPKCS1v15() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.args.plainText) {
				t.Errorf("RSA_DecryptPKCS1v15() = %v, want %v", got, tt.args.plainText)
			}
		})
	}

	// 加大原文长度
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				tt.args.plainText = append(tt.args.plainText, tt.args.plainText...)
			}
			newPlainText := tt.args.plainText
			cipherText, err := RSA_EncryptPKCS1v15(newPlainText, publicKey)
			if err != nil {
				t.Errorf("RSA_EncryptPKCS1v15() error = %v", err)
				return
			}
			got, err := RSA_DecryptPKCS1v15(cipherText, privateKey)
			if err != nil {
				t.Errorf("RSA_DecryptPKCS1v15() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, newPlainText) {
				t.Errorf("RSA_DecryptPKCS1v15() = %v, want %v", got, tt.args.plainText)
			}
		})
	}
}

func TestRsaSignPKCS1v15(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{[]byte("hello world")}},
		{"", args{[]byte("hello world, 你好，世界")}},
		{"", args{[]byte("gopher !!!")}},
	}
	privateKey, publicKey, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RsaSignPKCS1v15(tt.args.src, privateKey)
			if err != nil {
				t.Errorf("RsaSignPKCS1v15() error = %v", err)
				return
			}
			if err := RsaVerifyPKCS1v15(tt.args.src, got, publicKey); err != nil {
				t.Errorf("RsaVerifyPKCS1v15() error = %v", err)
			}
		})
	}
}
