# cnblogs-publisher

用 Go 编写的博客园命令行发布工具，通过 [MetaWeblog API](https://rpc.cnblogs.com/metaweblog/) 实现文章的发布、更新、查询和删除。

零第三方依赖，纯 Go 标准库实现，编译一次到处跑。

## 快速开始

### 1. 编译

```bash
cd cnblogs-publisher
go build -o cnblogs-publisher.exe .    # Windows
go build -o cnblogs-publisher .        # Linux / macOS
```

### 2. 配置

编辑 `etc/config.yaml`（和可执行文件放同一目录结构下，或通过 `--config` 指定路径）：

```yaml
# 博客园 MetaWeblog API 配置
blog_url: "https://rpc.cnblogs.com/metaweblog/你的博客名"
blog_id: "你的博客ID"
username: "你的用户名"
token: "你的MetaWeblog访问令牌"
```

| 字段 | 说明 |
|------|------|
| `blog_url` | `https://rpc.cnblogs.com/metaweblog/` + 你的博客地址名。例如博客地址是 `cnblogs.com/zhangsan`，则填 `https://rpc.cnblogs.com/metaweblog/zhangsan` |
| `blog_id` | 登录博客园后台，在浏览器地址栏可以看到 |
| `username` | 博客园登录用户名 |
| `token` | 博客园后台 → 设置 → 博客设置 → 其他设置 → MetaWeblog 访问令牌 |

### 3. 使用

#### 发布新文章

```bash
./cnblogs-publisher publish posts/my-article.md
```

Markdown 文件中第一个 `# 标题` 会自动提取为文章标题，其余内容作为正文。

#### 带分类和标签

```bash
./cnblogs-publisher publish posts/my-article.md --categories "Go,教程" --tags "golang,博客"
```

#### 保存为草稿

```bash
./cnblogs-publisher publish posts/my-article.md --draft
```

#### 更新已有文章

发布成功后会返回文章 ID，下次更新带上 `--postid`：

```bash
./cnblogs-publisher publish posts/my-article.md --postid 12345678
```

#### 列出最近文章

```bash
./cnblogs-publisher list        # 默认 10 篇
./cnblogs-publisher list 20     # 最近 20 篇
```

#### 删除文章

```bash
./cnblogs-publisher delete 12345678
```

#### 指定配置文件

所有命令都支持 `--config`：

```bash
./cnblogs-publisher publish posts/my-article.md --config ~/my-config.yaml
```

## Markdown 文章格式

```markdown
# 文章标题

正文内容，支持标准 Markdown 语法。

## 二级标题

代码、图片等都可以正常使用。
```

如果没有 `# ` 标题行，文章标题默认为"无标题"。

## 命令参考

```
cnblogs-publisher <命令> [参数]

命令:
  publish <文件>       发布或更新文章
  list [数量]          列出最近的文章（默认 10）
  delete <文章ID>      删除文章
  help                 显示帮助

publish 参数:
  --postid <ID>        更新已有文章
  --categories <分类>  逗号分隔
  --tags <标签>        逗号分隔
  --draft              保存为草稿
  --config <路径>      配置文件路径（默认 etc/config.yaml）
```

## 项目结构

```
cnblogs-publisher/
├── main.go                          # CLI 入口
├── etc/
│   └── config.yaml                  # 配置文件（已 gitignore）
├── posts/                           # Markdown 文章目录
├── internal/
│   ├── cnblogs/client.go            # 博客园 API 封装
│   └── xmlrpc/                      # 纯标准库 XML-RPC 实现
│       ├── client.go
│       ├── encode.go
│       └── decode.go
└── go.mod
```

## 注意事项

- `etc/config.yaml` 含敏感信息，已在 `.gitignore` 中排除
- 博客园后台需开启 Markdown 编辑器，发送的 Markdown 内容才能正常渲染
- HTTP 请求超时为 30 秒，网络异常时会自动报错而非挂起
