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

var httpClient = DefaultHTTPClient()

// Head 发送Head请求
func Head(args *RequestArgs) error {
	return httpClient.Head(args)
}

// Get 发送Get请求
func Get(args *RequestArgs) error {
	return httpClient.Get(args)
}

// Post 发送Post请求
func Post(args *RequestArgs) error {
	return httpClient.Post(args)
}

// Put 发送Put请求
func Put(args *RequestArgs) error {
	return httpClient.Put(args)
}

// Delete 发送Delete请求
// 注意：发送Delete请求的时候，不要使用Params携带参数，统一放置到Body中，否则远程可能收不到数据
func Delete(args *RequestArgs) error {
	return httpClient.Delete(args)
}

// ResetDefaultClient 替换默认的HTTP客户端
func ResetDefaultClient(c *HTTPClient) {
	httpClient = c
}

// SetDebug 设置Debug功能
// 如果logWriter非空，则启用debug日志输出，否则禁用
func SetDebug(w logWriter) {
	httpClient.Debug = w
}

type logWriter interface {
	Println(format string, v ...interface{})
}
