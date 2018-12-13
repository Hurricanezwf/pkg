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

package encoding

import (
	"bytes"
	"errors"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type EncodeMethod string

const (
	WithPB   = EncodeMethod("protobuf")
	WithJSON = EncodeMethod("json") // 不会缺省“零值”字段
)

func EncodeMsgBodyForMQ(msg proto.Message) ([]byte, error) {
	return Encode(WithPB, msg)
}

func DecodeMsgBodyFromMQ(msgBody []byte, msg proto.Message) error {
	return Decode(WithPB, msgBody, msg)
}

func EncodeFSMContext(ctx proto.Message) ([]byte, error) {
	return Encode(WithPB, ctx)
}

func DecodeFSMContext(b []byte, ctx proto.Message) error {
	return Decode(WithPB, b, ctx)
}

func Encode(method EncodeMethod, msg proto.Message) (b []byte, err error) {
	switch method {
	case WithPB:
		b, err = encodeWithPB(msg)
	case WithJSON:
		b, err = encodeWithJSON(msg)
	default:
		err = errors.New("Invalid EncodeMethod")
	}
	return b, err
}

func Decode(method EncodeMethod, msgBody []byte, msg proto.Message) (err error) {
	if len(msgBody) <= 0 {
		return errors.New("Empty msg body")
	}

	switch method {
	case WithPB:
		err = decodeWithPB(msgBody, msg)
	case WithJSON:
		err = decodeWithJSON(msgBody, msg)
	default:
		err = errors.New("Invalid EncodeMethod")
	}
	return err
}

func encodeWithPB(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func decodeWithPB(msgBody []byte, msg proto.Message) error {
	return proto.Unmarshal(msgBody, msg)
}

func encodeWithJSON(msg proto.Message) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	m := jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	if err := m.Marshal(buf, msg); err != nil {
		return nil, err
	}

	b := make([]byte, buf.Len())
	_, err := buf.Read(b)
	return b, err
}

func decodeWithJSON(msgBody []byte, msg proto.Message) error {
	buf := bytes.NewBuffer(msgBody)
	return jsonpb.Unmarshal(buf, msg)
}
