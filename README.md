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

2. **本地 skill**，让 agent 知道怎么用上面那个插件。装好：

   ```bash
   curl -fsSL https://kimi-web-img.moonshot.cn/webbridge/install.sh | bash
   ```

### 安装 skill

```bash
npx skills add better-world-ai/x-cli
```

<details>
<summary>没有 Node.js？手动安装</summary>

把 `skills/agent-cli-creator/` 复制到你 agent 的 skills 目录即可（Claude Code 是 `~/.claude/skills/`）。不确定路径？把这一段 README 丢给你的 agent，它会自己判断。

</details>

装完就能用，对话里说一句「帮我给 example.com 做个 CLI」即可触发。

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

去 [Releases 页面](https://github.com/better-world-ai/x-cli/releases) 下载对应平台的归档，解压即可用。

### macOS 打开提示

遇到「无法打开，因为开发者身份未验证」时，执行：

```bash
xattr -d com.apple.quarantine ./<cli-name>
```

### 本地编译

```bash
git clone https://github.com/better-world-ai/x-cli
cd x-cli/<某个-cli>
go build -o ./<cli-name> .
```

## License

MIT，见 [LICENSE](./LICENSE)。
