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

// FilterFunc 过滤器函数
type FilterFunc func(args *RequestArgs) error

// RequestArgs 通用请求参数封装
type RequestArgs struct {
	// URL 请求地址 (必填)
	URL string

	// Headers HTTP请求头设置 (可选)
	// nil表示不设置请求头
	Headers map[string]string

	// Params HTTP请求参数，键值对 (可选)
	Params map[string]string

	// Body HTTP请求体设置，必须是struct或[]byte类型 (可选)
	// nil表示无请求体
	Body interface{}

	// Filters 请求过滤器，会在请求发出前依次调用
	Filters []FilterFunc

	// JSONResult 接收JSON格式的响应内容, 必须是strcut类型 (可选)
	// 如果该字段非空，将自动解析至JSONResult
	JSONResult interface{}

	// BytesResult 接收字节流响应内容 (可选)
	// 如果该字段非空，响应体内容将被写入BytesResult
	BytesResult *bytes.Buffer
}

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

// complete 补全请求参数到BeegoHTTPRequest中
func (c *HTTPClient) complete(req *httplib.BeegoHTTPRequest, args *RequestArgs) error {
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

// filters 执行所有过滤器
func (c *HTTPClient) filters(args *RequestArgs) (err error) {
	for idx, f := range args.Filters {
		if err = f(args); err != nil {
			return fmt.Errorf("Call filter at index %d failed, %v", idx, err)
		}
	}
	return nil
}

// send 发送请求
func (c *HTTPClient) send(req *httplib.BeegoHTTPRequest, args *RequestArgs) error {
	var err error
	var rp *http.Response

	// 设置必要信息
	if err = c.complete(req, args); err != nil {
		return err
	}

	// 执行过滤器
	if err = c.filters(args); err != nil {
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
