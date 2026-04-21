# x-cli

基于 [kimi-webbridge](https://www.kimi.team/zh-cn/features/webbridge) 的自定义 CLI 模板。

装上之后，对你的 AI agent 说一句「帮我给 xxx 网站做个 CLI」，它会自动走完全部流程：分析站点 → 搭脚手架 → 实现命令 → 一边实现一边验证，最后产出一个可用的命令行工具（Go / Python / Node.js 都行）。

> 比如已有的 `baidu-cli` 就是这样生产出来的。

DEMO 示例：

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77


## 前置依赖

1. 先安装浏览器插件：
参考网站：
   - 中文：<https://www.kimi.com/zh-cn/features/webbridge>
   - English：<https://www.kimi.com/features/webbridge>

2. 交个 AI-Agent，一句话安装配置 kimi-webbridge：

```bash
curl -fsSL https://kimi-web-img.moonshot.cn/webbridge/install.sh | bash
```

## 安装

挑你在用的 agent：

### Claude Code

```bash
mkdir -p ~/.claude/skills
cp -r skills/agent-cli-creator ~/.claude/skills/
```

装完就能用，对话里说一句「帮我给 example.com 做个 CLI」即可触发。

### Kimi CLI

```bash
# 路径以你的 Kimi CLI 版本为准，可查 kimi-cli --help 或官方文档
cp -r skills/agent-cli-creator ~/.kimi/skills/
```

### OpenClaw

```bash
cp -r skills/agent-cli-creator <openclaw-的-skills-目录>/
```

如果 OpenClaw 不会自动加载，就在它的 agent 配置文件里加一条指向 `SKILL.md` 的引用。

### OpenAI Codex

Codex 读的是 `AGENTS.md`。把 `skills/agent-cli-creator/` 放在你的项目目录里，然后在 `AGENTS.md` 里加一段：

```md
## Skills

当用户要求为某个网站构建 CLI 时，请阅读并遵循：
`./skills/agent-cli-creator/SKILL.md`
```

## 怎么用

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
   每条命令都输出 JSON，方便串到别的脚本里。

## 目录

```
x-cli/
├── LICENSE
├── README.md
└── skills/
    └── agent-cli-creator/
        ├── SKILL.md                          # agent 主要读的指令
        └── references/                       # 各阶段的参考文档
            ├── site-exploration.md
            ├── login-handling.md
            ├── go-layout.md
            └── companion-skill-template.md
```

## License

MIT —— 见 [LICENSE](./LICENSE)。
