package httplib

import (
	"io"
	"time"
)

type HTTPSClient struct {
	// EnableDebug 是否启用调试输出请求和响应报文 (注：生产环境慎用)
	EnableDebug bool

	// DebugWriteTo 调试信息写入
	DebugWriteTo io.Writer

	// ConnectTimeout 连接超时
	ConnectTimeout time.Duration

	// RWTimeout 读写超时
	RWTimeout time.Duration

	// Retry 请求重试次数
	Retry int
}

func DefaultHTTPSClient() *HTTPSClient {
	return &HTTPSClient{
		EnableDebug:    false,
		DebugWriteTo:   nil,
		ConnectTimeout: 5 * time.Second,
		RWTimeout:      20 * time.Second,
		Retry:          2,
	}
}

func (c *HTTPSClient) Head(args *RequestArgs) error {
	// TODO
	return nil
}

func (c *HTTPSClient) Get(args *RequestArgs) error {
	// TODO
	return nil
}

func (c *HTTPSClient) Post(args *RequestArgs) error {
	// TODO
	return nil
}

func (c *HTTPSClient) Put(args *RequestArgs) error {
	// TODO
	return nil
}

func (c *HTTPSClient) Delete(args *RequestArgs) error {
	// TODO
	return nil
}
