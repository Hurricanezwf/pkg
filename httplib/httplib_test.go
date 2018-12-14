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

package httplib

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
)

type logwriter struct {
	l *log.Logger
}

func newLogWriter() logwriter {
	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	return logwriter{l: l}
}

func (w logwriter) Println(format string, v ...interface{}) {
	w.l.Output(2, fmt.Sprintf(format, v...))
}

func TestLog(t *testing.T) {
	l := newLogWriter()
	l.Println("hello, %s", "zwf")
	l.Println("hello, %s", "dsjk")
	l.Println("end")
}

func TestHTTPLib(t *testing.T) {
	SetDebug(newLogWriter())

	rp := bytes.NewBuffer(nil)
	err := Get(&RequestArgs{
		URL:         "http://www.google.cn",
		BytesResult: rp,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("Response Size: %dBytes\n", rp.Len())
}

func TestFilter(t *testing.T) {
	SetDebug(newLogWriter())

	filter0 := func(args *RequestArgs) error {
		return nil
	}
	filter1 := func(args *RequestArgs) error {
		return errors.New("rate limited")
	}

	rp := bytes.NewBuffer(nil)
	err := Get(&RequestArgs{
		URL:         "http://www.google.cn",
		Filters:     []FilterFunc{filter0, filter1},
		BytesResult: rp,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("Response Size: %dBytes\n", rp.Len())
}
