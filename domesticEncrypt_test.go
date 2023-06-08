package doraemon

import (
	"reflect"
	"testing"
)

func TestSM2Encrypt(t *testing.T) {
	type args struct {
		plainText []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{[]byte("hello world")}},
		{"", args{[]byte("hello world 你好世界")}},
		{"", args{[]byte("hello world 你好世界 こんにちは世界")}},
		{"", args{[]byte("hello world 你好世界 こんにちは世界 안녕하세요 세계")}},
	}
	pubKey, priKey, err := GenerateSM2Key()
	if err != nil {
		t.Errorf("GenerateSM2Key() error = %v", err)
		return
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SM2EncryptAsn1(pubKey, tt.args.plainText)
			if err != nil {
				t.Errorf("SM2Encrypt() error = %v", err)
				return
			}
			got1, err := SM2DecryptAsn1(priKey, got)
			if err != nil {
				t.Errorf("SM2Decrypt() error = %v", err)
				return
			}
			if !reflect.DeepEqual(tt.args.plainText, got1) {
				t.Errorf("SM2Decrypt() got1 = %v, want %v", got1, tt.args.plainText)
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
			got, err := SM2EncryptAsn1(pubKey, newPlainText)
			if err != nil {
				t.Errorf("SM2Encrypt() error = %v", err)
				return
			}
			got1, err := SM2DecryptAsn1(priKey, got)
			if err != nil {
				t.Errorf("SM2Decrypt() error = %v", err)
				return
			}
			if !reflect.DeepEqual(newPlainText, got1) {
				t.Errorf("SM2Decrypt() got1 = %v, want %v", got1, newPlainText)
			}
		})
	}
}
