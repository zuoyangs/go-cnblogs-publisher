package cnblogs

import (
	"bufio"
	"github.com/zuoyangs/go-cnblogs-publisher/internal/xmlrpc"
	"fmt"
	"os"
	"strings"
)

// Config holds the cnblogs MetaWeblog API credentials.
type Config struct {
	BlogURL  string
	BlogID   string
	Username string
	Token    string
}

func (c *Config) validate() error {
	if c.BlogURL == "" {
		return fmt.Errorf("config: blog_url 不能为空")
	}
	if c.Username == "" {
		return fmt.Errorf("config: username 不能为空")
	}
	if c.Token == "" {
		return fmt.Errorf("config: token 不能为空")
	}
	return nil
}

// parseYAML parses a simple flat key: value YAML file (no nesting, no arrays).
func parseYAML(data []byte) map[string]string {
	m := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		m[key] = val
	}
	return m
}

// Client wraps the MetaWeblog API calls.
type Client struct {
	rpc    *xmlrpc.Client
	config Config
}

// NewClient loads config from a YAML file and creates a ready-to-use client.
func NewClient(configPath string) (*Client, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	m := parseYAML(data)
	cfg := Config{
		BlogURL:  m["blog_url"],
		BlogID:   m["blog_id"],
		Username: m["username"],
		Token:    m["token"],
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &Client{
		rpc:    xmlrpc.NewClient(cfg.BlogURL),
		config: cfg,
	}, nil
}

// Post holds the data for publishing/editing an article.
type Post struct {
	Title      string
	Content    string
	Categories []string
	Tags       string
	Publish    bool
}

// buildPostStruct converts Post to an XML-RPC struct value.
func buildPostStruct(p Post) xmlrpc.Value {
	catValues := make([]xmlrpc.Value, len(p.Categories))
	for i, cat := range p.Categories {
		catValues[i] = xmlrpc.StringVal(cat)
	}
	return xmlrpc.StructVal([]xmlrpc.Member{
		{Name: "title", Value: xmlrpc.StringVal(p.Title)},
		{Name: "description", Value: xmlrpc.StringVal(p.Content)},
		{Name: "categories", Value: xmlrpc.ArrayVal(catValues)},
		{Name: "mt_keywords", Value: xmlrpc.StringVal(p.Tags)},
	})
}

// NewPost publishes a new post and returns the post ID.
func (c *Client) NewPost(p Post) (string, error) {
	resp, err := c.rpc.Call("metaWeblog.newPost", []xmlrpc.Param{
		{Value: xmlrpc.StringVal(c.config.BlogID)},
		{Value: xmlrpc.StringVal(c.config.Username)},
		{Value: xmlrpc.StringVal(c.config.Token)},
		{Value: buildPostStruct(p)},
		{Value: xmlrpc.BoolVal(p.Publish)},
	})
	if err != nil {
		return "", err
	}
	return xmlrpc.ExtractStringValue(resp), nil
}

// EditPost updates an existing post.
func (c *Client) EditPost(postID string, p Post) error {
	_, err := c.rpc.Call("metaWeblog.editPost", []xmlrpc.Param{
		{Value: xmlrpc.StringVal(postID)},
		{Value: xmlrpc.StringVal(c.config.Username)},
		{Value: xmlrpc.StringVal(c.config.Token)},
		{Value: buildPostStruct(p)},
		{Value: xmlrpc.BoolVal(p.Publish)},
	})
	return err
}

// PostInfo represents a blog post summary.
type PostInfo = xmlrpc.PostInfo

// GetRecentPosts returns recent posts.
func (c *Client) GetRecentPosts(count int) ([]PostInfo, error) {
	resp, err := c.rpc.Call("metaWeblog.getRecentPosts", []xmlrpc.Param{
		{Value: xmlrpc.StringVal(c.config.BlogID)},
		{Value: xmlrpc.StringVal(c.config.Username)},
		{Value: xmlrpc.StringVal(c.config.Token)},
		{Value: xmlrpc.IntVal(count)},
	})
	if err != nil {
		return nil, err
	}
	return xmlrpc.ExtractRecentPosts(resp), nil
}

// DeletePost deletes a post by ID.
func (c *Client) DeletePost(postID string) error {
	_, err := c.rpc.Call("blogger.deletePost", []xmlrpc.Param{
		{Value: xmlrpc.StringVal("cnblogs")},
		{Value: xmlrpc.StringVal(postID)},
		{Value: xmlrpc.StringVal(c.config.Username)},
		{Value: xmlrpc.StringVal(c.config.Token)},
		{Value: xmlrpc.BoolVal(true)},
	})
	return err
}
