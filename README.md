# x-cli

你想在网页上反复做的事，一句话告诉 AI agent，它就能帮你做成 CLI 工具。生成的 CLI 让 agent 随时调用，直接驱动你真实的 Chrome 登录态，不走 API，不折腾 token。

仓库里收录了几个这样做出来的 CLI，既能装好就用，也作为参考案例，演示 AI agent + [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge) 是怎么从一句需求生成一个完整 CLI 的。后文「自己做一个新 CLI」会走完整流程。

DEMO（一个 CLI 的诞生过程）：

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## 自己做一个新 CLI

仓库里几个 CLI 都是用 `skills/agent-cli-creator/` 这个 skill，让 AI agent 自动产出的。给你的 agent 装好下面这一套，对它说一句「帮我给 example.com 做个 CLI」就行。

### 前置依赖

要让 agent 真正控制你的浏览器，需要装 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)。它分两部分：

1. **浏览器插件**，agent 控制浏览器的入口工具。装好之后，所有点击、输入、读取都通过它转发，你登录过的 Chrome 会话自动被复用。
   - 中文：<https://www.kimi.com/zh-cn/features/webbridge>
   - English：<https://www.kimi.com/features/webbridge>

2. **本地 skill**，让 agent 知道怎么用上面那个插件。一行命令装好：

   ```bash
   curl -fsSL https://kimi-web-img.moonshot.cn/webbridge/install.sh | bash
   ```

### 安装 skill

挑你在用的 agent：

#### Claude Code

```bash
mkdir -p ~/.claude/skills
cp -r skills/agent-cli-creator ~/.claude/skills/
```

装完就能用，对话里说一句「帮我给 example.com 做个 CLI」即可触发。

#### Kimi CLI

```bash
cp -r skills/agent-cli-creator ~/.kimi/skills/
```

#### OpenClaw

```bash
cp -r skills/agent-cli-creator <openclaw-的-skills-目录>/
```

如果 OpenClaw 不会自动加载，就在它的 agent 配置文件里加一条指向 `SKILL.md` 的引用。

#### OpenAI Codex

Codex 读的是 `AGENTS.md`。把 `skills/agent-cli-creator/` 放在你的项目目录里，然后在 `AGENTS.md` 里加一段：

```md
## Skills

当用户要求为某个网站构建 CLI 时，请阅读并遵循：
`./skills/agent-cli-creator/SKILL.md`
```

### 怎么用

1. 启动 kimi-webbridge，并在 Chrome 里登录目标网站。
2. 对 agent 说，比如：
   > "帮我做一个 example.com 的 CLI，我要能拉首页信息流，并且能发评论。"
3. agent 会先问你几个问题（用什么语言、前 1–3 个功能是什么），然后自己去分析站点、搭脚手架、实现命令，关键节点会停下来确认。
4. 最终你会拿到一个这样用的工具：
   ```bash
   example-cli login-status
   example-cli home --limit 10
   example-cli post --content "hello"
   ```

## 包含的 CLI

| 工具 | 一句话 |
|---|---|
| [`baidu-cli`](./baidu-cli/) | 百度搜索，输出 JSON |
| [`google-cli`](./google-cli/) | Google 搜索 + 网页抓取，输出 JSON |
| [`nanobanana-cli`](./nanobanana-cli/) | 用 Gemini 2.5 Flash Image (Nano Banana) 生成图片 |
| [`chatgpt-image-cli`](./chatgpt-image-cli/) | 用 chatgpt.com/images 生成图片 |

## 安装预编译二进制

每个 CLI 的发布 tag 形如 `<cli-name>/v<version>`。在 [Releases 页面](https://github.com/better-world-ai/x-cli/releases) 找到你要的 CLI 最新 tag，然后：

```bash
# 以 google-cli v0.1.0 / macOS arm64 为例
TAG=google-cli/v0.1.0
curl -LO "https://github.com/better-world-ai/x-cli/releases/download/${TAG}/google-cli-darwin-arm64.tar.gz"
tar -xzf google-cli-darwin-arm64.tar.gz
./google-cli --help
```

每个 tag 都打包了 6 个平台的归档（约 3 MB / 个，gzip 压缩）：

| 平台 | 文件名后缀 |
|---|---|
| macOS arm64 (Apple Silicon) | `-darwin-arm64.tar.gz` |
| macOS amd64 (Intel) | `-darwin-amd64.tar.gz` |
| Linux amd64 | `-linux-amd64.tar.gz` |
| Linux arm64 (Graviton/树莓派 4+) | `-linux-arm64.tar.gz` |
| Windows amd64 | `-windows-amd64.zip` |
| Windows arm64 (Snapdragon 笔记本) | `-windows-arm64.zip` |

外加一份 `checksums.txt`（sha256）。

### macOS：解压后无法运行？

浏览器下载并解压后的文件带 `com.apple.quarantine` 标记，Gatekeeper 会拦：「无法打开，因为开发者身份未验证」。一行命令解除即可：

```bash
xattr -d com.apple.quarantine ./<cli-name>
```

### 本地编译

```bash
git clone https://github.com/better-world-ai/x-cli
cd x-cli/<某个-cli>
go build -o ./<cli-name> .
```

## 仓库结构

```
x-cli/
├── .github/workflows/
│   └── release.yml            # 统一的 per-CLI release workflow
├── skills/
│   └── agent-cli-creator/     # 用 AI agent 生成新 CLI 的 skill（见上文）
├── baidu-cli/                 # 独立项目
├── google-cli/                # 独立项目
├── nanobanana-cli/            # 独立项目
├── chatgpt-image-cli/         # 独立项目
├── LICENSE
└── README.md
```

每个 CLI 子目录是一个完整、独立的项目，自带依赖清单（如 `go.mod` / `pyproject.toml` / `package.json`）和 license 信息，可独立开发、独立发布。

## 发布流程

每个 CLI 用**带前缀的 tag** 触发独立发布，互不干扰：

```bash
git tag baidu-cli/v0.1.0          && git push origin baidu-cli/v0.1.0
git tag google-cli/v1.0.0         && git push origin google-cli/v1.0.0
git tag nanobanana-cli/v0.2.0     && git push origin nanobanana-cli/v0.2.0
git tag chatgpt-image-cli/v1.3.0  && git push origin chatgpt-image-cli/v1.3.0
```

CI 会自动识别 tag 前缀，只构建对应 CLI 的 6 个平台二进制（darwin arm64/amd64、linux amd64/arm64、windows amd64/arm64）并发布到 GitHub Release。也可以在 Actions 页面手动触发 workflow 做临时构建。

新增 Go CLI 时：在 `.github/workflows/release.yml` 的 `on.push.tags` 和 `workflow_dispatch.inputs.cli.options` 中加上对应名字。如果新 CLI 不是 Go（Python / TS 等），可以单独再加一个 sibling workflow（如 `release-python.yml`），用各自的 tag 前缀路由。

## License

MIT，见 [LICENSE](./LICENSE)。
