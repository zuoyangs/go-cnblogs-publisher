package xmlrpc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a minimal XML-RPC client.
type Client struct {
	URL        string
	httpClient *http.Client
}

// NewClient creates a Client with sensible defaults.
func NewClient(url string) *Client {
	return &Client{
		URL: url,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Call invokes an XML-RPC method and returns the raw response body.
func (c *Client) Call(method string, params []Param) ([]byte, error) {
	body := encodeRequest(method, params)

	resp, err := c.httpClient.Post(c.URL, "text/xml; charset=utf-8", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	if fault, ok := parseFault(data); ok {
		return nil, fmt.Errorf("博客园返回错误 [%d]: %s", fault.Code, fault.Message)
	}

	return data, nil
}

// Param types
type Param struct {
	Value Value
}

type Value struct {
	String *string
	Int    *int
	Bool   *bool
	Struct []Member
	Array  []Value
}

type Member struct {
	Name  string
	Value Value
}

// Helper constructors
func StringVal(s string) Value   { return Value{String: &s} }
func IntVal(i int) Value         { return Value{Int: &i} }
func BoolVal(b bool) Value       { return Value{Bool: &b} }
func StructVal(m []Member) Value { return Value{Struct: m} }
func ArrayVal(v []Value) Value   { return Value{Array: v} }
