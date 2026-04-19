package main

import (
	"github.com/zuoyangs/go-cnblogs-publisher/internal/cnblogs"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const usage = `博客园发布工具

用法:
  cnblogs-publisher <命令> [参数]

命令:
  publish <文件路径>   发布或更新文章
  list [数量]          列出最近的文章
  delete <文章ID>      删除文章

publish 参数:
  --postid <ID>        更新已有文章（不传则新建）
  --categories <分类>  分类，逗号分隔
  --tags <标签>        标签，逗号分隔
  --draft              保存为草稿（默认直接发布）
  --config <路径>      配置文件路径（默认 etc/config.yaml）

示例:
  cnblogs-publisher publish posts/my-article.md
  cnblogs-publisher publish posts/my-article.md --postid 12345678
  cnblogs-publisher publish posts/my-article.md --categories "Go,教程" --tags "golang"
  cnblogs-publisher list 20
  cnblogs-publisher delete 12345678
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "publish":
		runPublish(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "delete":
		runDelete(os.Args[2:])
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(1)
	}
}

func resolveConfigPath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	// 优先查找可执行文件同目录下的 etc/config.yaml
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "etc", "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "etc/config.yaml"
}

// extractTitle extracts the first H1 title from markdown content.
// Returns (title, bodyWithoutTitle).
func extractTitle(text string) (string, string) {
	lines := strings.SplitN(text, "\n", -1)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			// Remove the title line and rejoin
			remaining := append(lines[:i], lines[i+1:]...)
			body := strings.TrimSpace(strings.Join(remaining, "\n"))
			return title, body
		}
	}
	return "无标题", strings.TrimSpace(text)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func runPublish(args []string) {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	postID := fs.String("postid", "", "更新已有文章的 ID")
	categories := fs.String("categories", "", "分类，逗号分隔")
	tags := fs.String("tags", "", "标签，逗号分隔")
	draft := fs.Bool("draft", false, "保存为草稿")
	configPath := fs.String("config", "", "配置文件路径")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fatalf("请提供 Markdown 文件路径")
	}

	filePath := fs.Arg(0)
	raw, err := os.ReadFile(filePath)
	if err != nil {
		fatalf("❌ 读取文件失败: %v", err)
	}

	title, body := extractTitle(string(raw))

	client, err := cnblogs.NewClient(resolveConfigPath(*configPath))
	if err != nil {
		fatalf("❌ %v", err)
	}

	post := cnblogs.Post{
		Title:      title,
		Content:    body,
		Categories: splitCSV(*categories),
		Tags:       *tags,
		Publish:    !*draft,
	}

	if *postID != "" {
		if err := client.EditPost(*postID, post); err != nil {
			fatalf("❌ 更新失败: %v", err)
		}
		fmt.Printf("✅ 文章已更新 (ID: %s)\n", *postID)
	} else {
		newID, err := client.NewPost(post)
		if err != nil {
			fatalf("❌ 发布失败: %v", err)
		}
		fmt.Printf("✅ 文章已发布，ID: %s\n", newID)
		fmt.Printf("💡 下次更新: cnblogs-publisher publish %s --postid %s\n", filePath, newID)
	}
}

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	configPath := fs.String("config", "", "配置文件路径")
	fs.Parse(args)

	count := 10
	if fs.NArg() > 0 {
		fmt.Sscanf(fs.Arg(0), "%d", &count)
	}

	client, err := cnblogs.NewClient(resolveConfigPath(*configPath))
	if err != nil {
		fatalf("❌ %v", err)
	}

	posts, err := client.GetRecentPosts(count)
	if err != nil {
		fatalf("❌ 获取失败: %v", err)
	}

	fmt.Printf("📋 最近 %d 篇文章:\n\n", len(posts))
	for _, p := range posts {
		fmt.Printf("  [%s] %s\n", p.PostID, p.Title)
		if p.Link != "" {
			fmt.Printf("       %s\n", p.Link)
		}
		fmt.Println()
	}
}

func runDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	configPath := fs.String("config", "", "配置文件路径")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fatalf("请提供文章 ID")
	}

	postID := fs.Arg(0)
	client, err := cnblogs.NewClient(resolveConfigPath(*configPath))
	if err != nil {
		fatalf("❌ %v", err)
	}

	if err := client.DeletePost(postID); err != nil {
		fatalf("❌ 删除失败: %v", err)
	}
	fmt.Printf("✅ 文章已删除 (ID: %s)\n", postID)
}
