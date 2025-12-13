package mhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// FetchOptions 请求选项
type FetchOptions struct {
	URL        string
	Data       []byte
	DataMap    map[string]any
	Params     map[string]string
	Headers    map[string]string
	Timeout    int    // seconds
	Retry      int    // 重试次数
	RetryDelay int    // 重试次数延迟 seconds
	Method     string // 允许值：GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS（不区分大小写，会在 Do 中规范化为大写）
	// MaxBodySize 限制读取响应体的最大字节数，0 表示不限制
	MaxBodySize int64
	// Proxy 可选：为本次请求使用的代理地址，例如 "http://127.0.0.1:7890"
	Proxy string
}

// Fetch 请求封装
type Fetch struct {
	opts FetchOptions
}

// NewFetch 创建一个 Fetch 实例
func NewFetch(opts FetchOptions) *Fetch {
	return &Fetch{opts: opts}
}

// Get 发起 GET 请求，并返回响应 body
// Do 发起请求，使用 FetchOptions.Method，要求 Method 非空并且为标准 HTTP 方法
// 调用示例：
//
//	res, err := NewFetch(FetchOptions{URL: "https://...", Method: http.MethodPost, DataMap: m}).Do()
func (f *Fetch) Do() ([]byte, error) {
	opts := f.opts
	if opts.Method == "" {
		return nil, errors.New("empty Method")
	}
	m := strings.ToUpper(opts.Method)
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		// ok
	default:
		return nil, fmt.Errorf("invalid method: %s", opts.Method)
	}
	opts.Method = m
	return f.do(opts)
}

// do 执行请求，并支持重试
func (f *Fetch) do(opts FetchOptions) ([]byte, error) {
	// 保护性检查
	if opts.URL == "" {
		return nil, errors.New("empty URL")
	}

	// 构造 URL 和 params
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	if len(opts.Params) > 0 {
		q := u.Query()
		for k, v := range opts.Params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	// body 构造
	var rawBody []byte
	if opts.Method != http.MethodGet {
		if len(opts.Data) > 0 {
			rawBody = opts.Data
		} else if opts.DataMap != nil {
			jb, jerr := json.Marshal(opts.DataMap)
			if jerr != nil {
				return nil, jerr
			}
			rawBody = jb
			// 设置默认 Content-Type（修改本地 opts 副本，不影响调用方）
			if opts.Headers == nil {
				opts.Headers = map[string]string{"Content-Type": "application/json"}
			} else {
				if _, ok := opts.Headers["Content-Type"]; !ok {
					opts.Headers["Content-Type"] = "application/json"
				}
			}
		}
	}

	// 超时时间
	tout := 30
	if opts.Timeout > 0 {
		tout = opts.Timeout
	}

	// 重试参数
	retry := opts.Retry
	if retry < 0 {
		retry = 0
	}
	retryDelay := opts.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 1
	}

	// 使用 Colly 进行请求实现。Colly 会在内部处理连接重用、代理、超时等。
	var lastErr error

	// 组装 headers
	hdr := http.Header{}
	for k, v := range opts.Headers {
		hdr.Set(k, v)
	}

	for attempt := 0; attempt <= retry; attempt++ {
		// 创建 Collector
		c := colly.NewCollector()
		c.SetRequestTimeout(time.Duration(tout) * time.Second)

		if opts.Proxy != "" {
			if perr := c.SetProxy(opts.Proxy); perr != nil {
				return nil, fmt.Errorf("invalid proxy url: %w", perr)
			}
		}

		var respBody []byte
		var respStatus int
		var respErr error

		c.OnResponse(func(r *colly.Response) {
			respStatus = r.StatusCode
			if opts.MaxBodySize > 0 && int64(len(r.Body)) > opts.MaxBodySize {
				respBody = r.Body[:opts.MaxBodySize]
			} else {
				respBody = r.Body
			}
		})

		c.OnError(func(r *colly.Response, e error) {
			if r != nil {
				respStatus = r.StatusCode
				if r.Body != nil {
					// 限制大小
					if opts.MaxBodySize > 0 && int64(len(r.Body)) > opts.MaxBodySize {
						respErr = fmt.Errorf("http status %d: %s", r.StatusCode, string(r.Body[:opts.MaxBodySize]))
					} else {
						respErr = fmt.Errorf("http status %d: %s", r.StatusCode, string(r.Body))
					}
				} else {
					respErr = fmt.Errorf("http status %d: %w", r.StatusCode, e)
				}
			} else {
				respErr = e
			}
		})

		var bodyReader io.Reader
		if rawBody != nil {
			bodyReader = bytes.NewReader(rawBody)
		}

		// 发起请求（同步）
		reqErr := c.Request(opts.Method, u.String(), bodyReader, nil, hdr)
		if reqErr != nil {
			lastErr = fmt.Errorf("request error: %w", reqErr)
			if attempt < retry {
				time.Sleep(time.Duration(retryDelay) * time.Second)
				continue
			}
			return nil, lastErr
		}

		if respErr != nil {
			lastErr = respErr
			// 对 5xx 做重试
			if respStatus >= 500 && attempt < retry {
				time.Sleep(time.Duration(retryDelay) * time.Second)
				continue
			}
			return nil, lastErr
		}

		// 若没有通过 OnError 报错，但状态非 2xx，也视为错误
		if respStatus < 200 || respStatus >= 300 {
			lastErr = fmt.Errorf("http status %d: %s", respStatus, string(respBody))
			if respStatus >= 500 && attempt < retry {
				time.Sleep(time.Duration(retryDelay) * time.Second)
				continue
			}
			return nil, lastErr
		}

		return respBody, nil
	}

	return nil, lastErr
}
