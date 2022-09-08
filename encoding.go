package doraemon

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/axgle/mahonia"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

//GBK,GB18030
func Utf8ToGbk(text string) string {
	return mahonia.NewEncoder("gbk").ConvertString(text)
}

//GB18030
func Utf8ToANSI(text string) string {
	return mahonia.NewEncoder("GB18030").ConvertString(text)
}

func GbkToUtf8(b []byte) []byte {
	tfr := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(tfr)
	if e != nil {
		return nil
	}
	return d
}

//自动检测html编码,不会减少缓冲器的内容,
//得到编码器后用transform.NewReader解码为utf-8
func DetermineEncodingbyPeek(r *bufio.Reader, contentType string) (encoding.Encoding, error) {
	tempbytes, err := r.Peek(1024)
	if err != nil && err != io.EOF {
		return nil, err
	}
	e, _, _ := charset.DetermineEncoding(tempbytes, contentType)
	return e, nil
}

//转码utf-8,contentType 可以传空值""
//DetermineEncoding determines the encoding of an HTML document by
//examining up to the first 1024 bytes of content and the declared Content-Type.
func TransformCoding(rd io.Reader, contentType string) (*transform.Reader, error) {
	bodyReader := bufio.NewReader(rd)
	e, err := DetermineEncodingbyPeek(bodyReader, contentType)
	if err != nil {
		return nil, err
	}
	return transform.NewReader(bodyReader, e.NewDecoder()), nil
}

//对Unicode进行转码
func UnicodeStrToUtf8(str string) string {
	reg := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	return reg.ReplaceAllStringFunc(str, func(s string) string {
		r, _ := strconv.ParseInt(s[2:], 16, 32)
		return string(rune(r))
	})
}

//对Unicode进行转码
func UnicodeToUtf8(str []byte) []byte {
	reg := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	return reg.ReplaceAllFunc(str, func(s []byte) []byte {
		r, _ := strconv.ParseInt(string(s[2:]), 16, 32)
		return []byte(string(rune(r)))
	})
}
