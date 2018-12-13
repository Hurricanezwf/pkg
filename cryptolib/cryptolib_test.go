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
	"encoding/base64"
	"testing"
)

var (
	key       = []byte{0x63, 0x78, 0x79, 0x26, 0x64, 0x7a, 0x78, 0x26, 0x7a, 0x77, 0x66, 0x40, 0x32, 0x30, 0x31, 0x38}
	toEncrypt = []byte("Hello World, This is a secret, I won't tell anyone")
)

func TestAES256(t *testing.T) {
	encrypted, err := Encrypt(key, toEncrypt, TypeAES256)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("AES256: %s\n", base64.StdEncoding.EncodeToString(encrypted))

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err.Error())
	}

	if bytes.Equal(decrypted, toEncrypt) == false {
		t.Fatal("Not Equal")
	}
	t.Log("Success")
}

func TestAES128(t *testing.T) {
	encrypted, err := Encrypt(key, toEncrypt, TypeAES128)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("AES128: %s\n", base64.StdEncoding.EncodeToString(encrypted))

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err.Error())
	}

	if bytes.Equal(decrypted, toEncrypt) == false {
		t.Fatal("Not Equal")
	}
	t.Log("Success")
}

func TestXORBase64(t *testing.T) {
	encrypted, err := Encrypt(key, toEncrypt, TypeXORBase64)
	if err != nil {
		t.Fatal(err.Error())
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err.Error())
	}

	if bytes.Equal(decrypted, toEncrypt) == false {
		t.Fatal("Not Equal")
	}
	t.Log("Success")
}
