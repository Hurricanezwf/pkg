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

package mqwrapper

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	msg := []byte("This is request that create host")
	e := DefaultEncoder()

	// 编码
	b, err := e.Encode(10130, msg)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("After Encode: %#v\n", b)

	// 解码
	actionKey, msgBody, err := e.Decode(b)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("ActionKey: %s\n", string(actionKey))
	fmt.Printf("After Decode: %#v\n", msgBody)

	if bytes.Compare(msg, msgBody) != 0 {
		t.Fatal("Not equal")
	}
}

func TestCompress(t *testing.T) {
	msg := []byte("Hello, My name is")

	compressed, err := Compress(msg)
	if err != nil {
		t.Fatalf("Compress error: %v\n", err)
	}

	b, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress error: %v\n", err)
	}

	if bytes.Compare(msg, b) != 0 {
		t.Fatalf("Not equal\n")
	}
}
