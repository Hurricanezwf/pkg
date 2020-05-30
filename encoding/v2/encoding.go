package encodingv2

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

// Encoding 编解码器的抽象
type Encoding interface {
	EncodeTo(w io.Writer, v interface{}) error
	DecodeFrom(r io.Reader, v interface{}) error
}

func DefaultEncoding() Encoding {
	return NewProtoEncoding()
}

// GobEncoding 使用Gob的方式编解码
type GobEncoding struct{}

func NewGobEncoding() *GobEncoding {
	return &GobEncoding{}
}

func (e *GobEncoding) EncodeTo(w io.Writer, v interface{}) error {
	return gob.NewEncoder(w).Encode(v)
}

func (e *GobEncoding) DecodeFrom(r io.Reader, v interface{}) error {
	return gob.NewDecoder(r).Decode(v)
}

// ProtoEncoding 使用protobuf的方式编解码
type ProtoEncoding struct{}

func NewProtoEncoding() *ProtoEncoding {
	return &ProtoEncoding{}
}

func (e *ProtoEncoding) EncodeTo(w io.Writer, v interface{}) error {
	if _, ok := v.(proto.Message); !ok {
		return errors.New("v must be type of proto.Message")
	}
	if b, err := proto.Marshal(v.(proto.Message)); err != nil {
		return err
	} else {
		w.Write(b)
	}
	return nil
}

func (e *ProtoEncoding) DecodeFrom(r io.Reader, v interface{}) error {
	if _, ok := v.(proto.Message); !ok {
		return errors.New("v must be type of proto.Message")
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, v.(proto.Message))
}

// JsonEncoding 使用json的方式编码
type JsonEncoding struct{}

func NewJsonEncoding() *JsonEncoding {
	return &JsonEncoding{}
}

func (e *JsonEncoding) EncodeTo(w io.Writer, v interface{}) error {
	if b, err := json.Marshal(v); err != nil {
		return err
	} else {
		w.Write(b)
	}
	return nil
}

func (e *JsonEncoding) DecodeFrom(r io.Reader, v interface{}) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
