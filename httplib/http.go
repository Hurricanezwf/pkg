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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
)

type HTTPClient struct {
	// EnableHTTPS 是否启用HTTPS
	EnableHTTPS bool

	// TLSConfig 使用HTTPS的配置
	TLSConfig *tls.Config

	// ConnectTimeout 连接超时
	ConnectTimeout time.Duration

	// RWTimeout 读写超时
	RWTimeout time.Duration

	// Retry 请求重试次数
	Retry int

	// Debug 调试信息写入
	Debug logWriter
}

func DefaultHTTPClient() *HTTPClient {
	return &HTTPClient{
		EnableHTTPS:    false,
		TLSConfig:      nil,
		ConnectTimeout: 5 * time.Second,
		RWTimeout:      20 * time.Second,
		Retry:          1,
		Debug:          nil,
	}
}

func (c *HTTPClient) Head(args *RequestArgs) error {
	return c.send(httplib.Head(args.URL), args)
}

func (c *HTTPClient) Get(args *RequestArgs) error {
	return c.send(httplib.Get(args.URL), args)
}

func (c *HTTPClient) Post(args *RequestArgs) error {
	return c.send(httplib.Post(args.URL), args)
}

func (c *HTTPClient) Put(args *RequestArgs) error {
	return c.send(httplib.Put(args.URL), args)
}

func (c *HTTPClient) Delete(args *RequestArgs) error {
	return c.send(httplib.Delete(args.URL), args)
}

func (c *HTTPClient) prepare(req *httplib.BeegoHTTPRequest, args *RequestArgs) error {
	// 设置请求头
	for headerK, headerV := range args.Headers {
		if strings.ToLower(headerK) != "host" {
			req.Header(headerK, headerV)
		} else {
			req.SetHost(headerV)
		}
	}

	// 设置请求参数
	for reqK, reqV := range args.Params {
		req.Param(reqK, reqV)
	}

	// 设置是否启用HTTPS
	if c.EnableHTTPS {
		req.SetTLSClientConfig(c.TLSConfig)
	}

	// 设置超时时间
	req.SetTimeout(c.ConnectTimeout, c.RWTimeout)

	// 设置重试次数
	req.Retries(c.Retry)

	// 设置Debug
	req.Debug((c.Debug != nil))

	// 设置请求体
	if args.Body == nil {
		return nil
	}
	b, ok := args.Body.([]byte)
	if !ok {
		var err error
		if b, err = json.Marshal(args.Body); err != nil {
			return fmt.Errorf("Marshal body failed, %v", err)
		}
	}
	if len(b) > 0 {
		req.Body(b)
	}
	return nil
}

func (c *HTTPClient) send(req *httplib.BeegoHTTPRequest, args *RequestArgs) error {
	var err error
	var rp *http.Response

	// 设置必要信息
	if err = c.prepare(req, args); err != nil {
		return err
	}

	// 发送请求
	if rp, err = req.Response(); err != nil {
		return err
	}
	defer rp.Body.Close()

	if c.Debug != nil {
		c.Debug.Println("\n%s", string(req.DumpRequest()))
	}

	// 读取响应体
	var n int64
	var buf = bytes.NewBuffer(nil)
	if n, err = buf.ReadFrom(rp.Body); err != nil {
		return fmt.Errorf("Read http body failed, %v", err)
	}

	// 解析结果
	if rp.StatusCode != 200 {
		return fmt.Errorf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
	}
	if args.JSONResult != nil {
		if c.Debug != nil {
			c.Debug.Println("[%d Bytes] %s: %s\n", buf.Len(), reflect.TypeOf(args.JSONResult), buf.String())
		}
		if err = json.Unmarshal(buf.Next(int(n)), args.JSONResult); err != nil {
			return fmt.Errorf("Bad response format, %v", err)
		}
	}
	if args.BytesResult != nil {
		if c.Debug != nil {
			c.Debug.Println("[%d Bytes] %s", buf.Len(), buf.String())
		}
		buf.WriteTo(args.BytesResult)
	}
	return nil
}
