package doraemon

import "testing"

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
}
