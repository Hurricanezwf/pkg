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
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
)

type MsgEncoder interface {
	Encode(actionKey int32, msgBody []byte) ([]byte, error)
	Decode(b []byte) (actionKey int32, msgBody []byte, err error)
}

func DefaultEncoder() MsgEncoder {
	return newMsgEncoderV1()
}

type msgEncoderV1 struct {
	// 消息魔数
	magicN byte

	// 最大报文段长度
	maxSegmentLen uint32
}

func newMsgEncoderV1() MsgEncoder {
	return &msgEncoderV1{
		magicN:        0x22,
		maxSegmentLen: 5242880, // 默认限制5MB
	}
}

// Encode 编码MQ消息
// 1 Byte : 魔数
// 3 Bytes: 编码选项
// 4 Bytes: ActionKey
// 4 Bytes: MsgBodyLen
// N Bytes: MsgBody
//
// 编码选项(从左至右分别为bit0~bit24)：
// bit0表示msgBody是否压缩，1表示压缩，0表示不压缩
// bit1~bit23暂时预留
//
func (e *msgEncoderV1) Encode(actionKey int32, msgBody []byte) ([]byte, error) {
	// 超过50KB的消息体，将进行gzip压缩，压缩级别为5
	var err error
	var isCompressed bool
	if len(msgBody) > 51200 {
		msgBody, err = Compress(msgBody)
		if err != nil {
			return nil, fmt.Errorf("Compress msg body failed, %v", err)
		}
		isCompressed = true
	}

	// 消息长度限制
	segmentLen := 1 + 3 + 4 + 4 + len(msgBody)
	if uint32(segmentLen) > e.maxSegmentLen {
		return nil, errors.New("msg too long")
	}

	offset := 0
	buf := make([]byte, segmentLen)

	// MagicNumber
	buf[offset] = e.magicN

	// Header Option
	offset++
	if isCompressed {
		buf[offset] = 0x80
	}

	// ActionKey
	offset += 3
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(actionKey))

	// MsgBodyLen
	offset += 4
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(len(msgBody)))

	// MsgBody
	offset += 4
	copy(buf[offset:offset+len(msgBody)], msgBody)

	return buf, nil
}

// Decode 解码MQ消息
func (e *msgEncoderV1) Decode(b []byte) (action int32, msgBody []byte, err error) {
	// 验证长度
	if uint32(len(b)) > e.maxSegmentLen {
		err = errors.New("msg is too long")
		return
	}
	if len(b) < 12 {
		err = errors.New("msg too short")
		return
	}

	var offset uint32 = 0
	var isCompressed bool

	// 验证魔数
	if b[0] != e.magicN {
		err = errors.New("bad msg format, magicN didn't match")
		return
	}

	// 获取头部选项
	if b[1]&0x80 > 0 {
		isCompressed = true
	}

	// 解析ActionKey
	offset = 4
	action = int32(binary.BigEndian.Uint32(b[offset : offset+4]))

	// 解析MsgBody长度
	offset += 4
	msgBodyLen := binary.BigEndian.Uint32(b[offset : offset+4])

	// 解析MsgBody
	offset += 4
	bodyEnd := offset + msgBodyLen
	if bodyEnd > uint32(len(b)) {
		err = errors.New("msg too short")
		return
	}
	msgBody = b[offset:bodyEnd]

	// 解压缩
	if isCompressed {
		msgBody, err = Decompress(msgBody)
		if err != nil {
			err = fmt.Errorf("Decompress msg body failed, %v", err)
			return
		}
	}

	return action, msgBody, nil
}

func Compress(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	w, err := gzip.NewWriterLevel(b, 5)
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data); err != nil {
		return nil, err
	}
	w.Close() // 这里的close不作为资源清理，而是将缓存中的数据刷入bytes.Buffer
	return b.Bytes(), nil
}

func Decompress(compressed []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}
