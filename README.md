# x-cli

[autoclaw-cc](https://github.com/autoclaw-cc) 旗下 CLI 工具的 monorepo。每个 CLI 都是一个独立项目（Go / Python / TS 都行——目前都是 Go），由 AI agent + [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge) 自动生成；它们驱动你**真实的 Chrome 登录会话**，不走 API、不折腾 token，直接复用现成的网页登录态。

> 例如下面的 `baidu-cli`、`google-cli` 等都是这样产出的——后文「自己做一个新 CLI」演示完整流程。

## 包含的 CLI

| 工具 | 一句话 |
|---|---|
| [`baidu-cli`](./baidu-cli/) | 百度搜索，输出 JSON |
| [`google-cli`](./google-cli/) | Google 搜索 + 网页抓取，输出 JSON |
| [`nanobanana-cli`](./nanobanana-cli/) | 用 Gemini 2.5 Flash Image (Nano Banana) 生成图片 |
| [`chatgpt-image-cli`](./chatgpt-image-cli/) | 用 chatgpt.com/images 生成图片 |

DEMO 示例：

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## 安装预编译二进制（推荐）

每个 CLI 的发布 tag 形如 `<cli-name>/v<version>`。在 [Releases 页面](https://github.com/autoclaw-cc/x-cli/releases) 找到你要的 CLI 最新 tag，然后：

```bash
# 以 google-cli v1.0.0 / macOS arm64 为例
TAG=google-cli/v1.0.0
curl -L -o google-cli \
  "https://github.com/autoclaw-cc/x-cli/releases/download/${TAG}/google-cli-darwin-arm64"
chmod +x google-cli
./google-cli --help
```

每个 tag 都包含 4 个平台的二进制：`<cli>-darwin-arm64`、`<cli>-darwin-amd64`、`<cli>-linux-amd64`、`<cli>-windows-amd64.exe`，外加一份 `checksums.txt`（sha256）。

### macOS：从浏览器下载后无法运行？

浏览器下载的文件会被 macOS 加上 `com.apple.quarantine` 标记，运行时会被拦下来「无法打开，因为开发者身份未验证」。一行命令解除即可：

```bash
xattr -d com.apple.quarantine ./<cli-name>
```

### 本地编译

```bash
git clone https://github.com/autoclaw-cc/x-cli
cd x-cli/<某个-cli>
go build -o ./<cli-name> .
```

## 仓库结构

```
x-cli/
├── .github/workflows/
│   └── release.yml            # 统一的 per-CLI release workflow
├── skills/
│   └── agent-cli-creator/     # 用 AI agent 生成新 CLI 的 skill（见下文）
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

CI 会自动识别 tag 前缀，只构建对应 CLI 的 4 个平台二进制（darwin arm64/amd64、linux amd64、windows amd64）并发布到 GitHub Release。也可以在 Actions 页面手动触发 workflow 做临时构建。

新增 Go CLI 时：在 `.github/workflows/release.yml` 的 `on.push.tags` 和 `workflow_dispatch.inputs.cli.options` 中加上对应名字。如果新 CLI 不是 Go（Python / TS 等），可以单独再加一个 sibling workflow（如 `release-python.yml`），用各自的 tag 前缀路由。

## 自己做一个新 CLI

仓库里几个 CLI 都是用 `skills/agent-cli-creator/` 这个 skill 让 AI agent 自动产出的。

### 前置依赖

1. 先安装浏览器插件：
   - 中文：<https://www.kimi.com/zh-cn/features/webbridge>
   - English：<https://www.kimi.com/features/webbridge>

2. 交个 AI-Agent，一句话安装配置 kimi-webbridge：

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

## License

MIT —— 见 [LICENSE](./LICENSE)。
