// Copyright 2018 Hurricanezwf. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cryptolib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	TypeXORBase64 byte = 0x01
	TypeAES128    byte = 0x02
	TypeAES256    byte = 0x03
)

func Encrypt(key, toEncrypt []byte, encType byte) ([]byte, error) {
	var b []byte
	var err error

	switch encType {
	case TypeXORBase64:
		b, err = EncryptWithXORBase64(key, toEncrypt)
	case TypeAES128:
		b, err = EncryptWithAES128(key, toEncrypt)
	case TypeAES256:
		b, err = EncryptWithAES256(key, toEncrypt)
	default:
		return nil, errors.New("No Encrypt method found")
	}

	if err == nil {
		wrap := make([]byte, len(b)+1)
		wrap[len(b)] = encType
		copy(wrap, b)
		return wrap, nil
	}
	return nil, err
}

func Decrypt(key, toDecrypt []byte) ([]byte, error) {
	if len(toDecrypt) < 1 {
		return nil, errors.New("Bad content to decrypt")
	}
	t := toDecrypt[len(toDecrypt)-1]
	switch t {
	case TypeXORBase64:
		return DecryptWithXORBase64(key, toDecrypt[:len(toDecrypt)-1])
	case TypeAES128:
		return DecryptWithAES128(key, toDecrypt[:len(toDecrypt)-1])
	}
	return nil, errors.New("No Decrypt method found")
}

func EncryptWithAES256(key, src []byte) ([]byte, error) {
	return aesEncrypt(key, src, 256)
}

func DecryptWithAES256(key, src []byte) ([]byte, error) {
	return aesDecrypt(key, src, 256)
}

func EncryptWithAES128(key, src []byte) ([]byte, error) {
	return aesEncrypt(key, src, 128)
}

func DecryptWithAES128(key, src []byte) ([]byte, error) {
	return aesDecrypt(key, src, 128)
}

func aesEncrypt(key, src []byte, bit int) ([]byte, error) {
	// add padding
	toEncrypt := make([]byte, 0, len(src))
	toEncrypt = append(toEncrypt, src...)
	toEncrypt = PKCS7Padding(toEncrypt, aes.BlockSize)
	if len(toEncrypt)%aes.BlockSize != 0 {
		return nil, errors.New("Content to encrypt is not a multiple of the block size")
	}

	k := make([]byte, 0, 32)
	k = append(k, key...)
	block, err := aes.NewCipher(makeKey(k, bit/8))
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(toEncrypt))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], toEncrypt)
	return ciphertext, nil
}

func aesDecrypt(key, src []byte, bit int) ([]byte, error) {
	if len(src) < aes.BlockSize {
		return nil, errors.New("Content to decrypt to short")
	}

	k := make([]byte, 0, 32)
	k = append(k, key...)
	block, err := aes.NewCipher(makeKey(k, bit/8))
	if err != nil {
		return nil, err
	}

	toDecrypt := make([]byte, 0, len(src))
	toDecrypt = append(toDecrypt, src...)
	iv := toDecrypt[:aes.BlockSize]
	toDecrypt = toDecrypt[aes.BlockSize:]
	if len(toDecrypt)%aes.BlockSize != 0 {
		return nil, errors.New("Content to decrypt is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(toDecrypt, toDecrypt)
	toDecrypt = PKCS7UnPadding(toDecrypt)
	return toDecrypt, nil
}

// XORBase64
func EncryptWithXORBase64(key, src []byte) ([]byte, error) {
	k := make([]byte, 0, len(src))
	tmpSrc := make([]byte, len(src))
	k = append(k, key...)
	k = makeKey(k, len(src))
	for i, b := range src {
		tmpSrc[i] = k[i] ^ b
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(tmpSrc)))
	base64.StdEncoding.Encode(dst[:], tmpSrc)
	return dst, nil
}

func DecryptWithXORBase64(key, src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	dst = dst[:n]
	k := make([]byte, 0, len(dst))
	k = append(k, key...)
	k = makeKey(k, len(dst))
	for i, b := range dst {
		dst[i] = k[i] ^ b
	}
	return dst, nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func makeKey(key []byte, size int) []byte {
	for len(key) < size {
		key = append(key, key...)
	}
	return key[0:size]
}
