# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

English | [中文](./README.md)

Tell an AI agent in one sentence what you keep doing on a webpage, and it'll turn that into a CLI tool. The generated CLI can be called by your agent any time, driving your real Chrome login session directly — no API, no token juggling.

This repo collects several CLIs built exactly this way. They're installable and useful as-is, but they also serve as reference cases showing how AI agent + [kimi-webbridge](https://www.kimi.com/features/webbridge) turns a one-line request into a complete CLI. The "Build your own CLI" section below walks through the full flow.

DEMO (a CLI being born):

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## Build your own CLI

All the CLIs in this repo were produced automatically by AI agents using the `skills/agent-cli-creator/` skill. Set your agent up with the toolchain below, then say "Build me a CLI for example.com" — that's it.

### Prerequisites

To let the agent actually drive your browser, install [kimi-webbridge](https://www.kimi.com/features/webbridge). It has two parts:

1. **Browser extension** — the agent's entry point for controlling the browser. Once installed, every click, input, and read is forwarded through it, and your existing Chrome login sessions get reused automatically.
   - English: <https://www.kimi.com/features/webbridge>
   - 中文：<https://www.kimi.com/zh-cn/features/webbridge>

2. **Local skill** that teaches the agent how to use the extension above. Install:

   ```bash
   curl -fsSL https://kimi-web-img.moonshot.cn/webbridge/install.sh | bash
   ```

### Install the skill

```bash
npx skills add better-world-ai/x-cli
```

<details>
<summary>No Node.js? Manual install</summary>

Copy `skills/agent-cli-creator/` into your agent's skills directory (for Claude Code that's `~/.claude/skills/`). Not sure where it goes? Paste this README section to your agent — it'll figure it out.

</details>

Once installed, just say "Build me a CLI for example.com" in conversation to trigger it.

### How to use

1. Start kimi-webbridge and log into the target site in Chrome.
2. Tell your agent, e.g.:
   > "Build me a CLI for example.com. I want to pull the homepage feed and post comments."
3. The agent asks a few questions first (which language, what the first 1–3 features are), then goes off to analyze the site, scaffold the project, and implement the commands — pausing at key checkpoints to confirm with you.
4. You end up with a tool used like this:
   ```bash
   example-cli login-status
   example-cli home --limit 10
   example-cli post --content "hello"
   ```

## Included CLIs

| Tool | One-liner |
|---|---|
| [`baidu-cli`](./baidu-cli/) | Baidu search, JSON output |
| [`google-cli`](./google-cli/) | Google search + page scraping, JSON output |
| [`nanobanana-cli`](./nanobanana-cli/) | Generate images with Gemini 2.5 Flash Image (Nano Banana) |
| [`chatgpt-image-cli`](./chatgpt-image-cli/) | Generate images via chatgpt.com/images |

## Install prebuilt binaries

Grab the archive for your platform from the [Releases page](https://github.com/better-world-ai/x-cli/releases), extract it, and run.

### macOS Gatekeeper prompt

If you see "cannot be opened because the developer cannot be verified", run:

```bash
xattr -d com.apple.quarantine ./<cli-name>
```

### Build from source

```bash
git clone https://github.com/better-world-ai/x-cli
cd x-cli/<some-cli>
go build -o ./<cli-name> .
```

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/xpzouying"><img src="https://avatars.githubusercontent.com/u/3946563?v=4?s=100" width="100px;" alt="zy"/><br /><sub><b>zy</b></sub></a><br /><a href="https://github.com/better-world-ai/x-cli/commits?author=xpzouying" title="Code">💻</a> <a href="https://github.com/better-world-ai/x-cli/commits?author=xpzouying" title="Documentation">📖</a> <a href="#maintenance-xpzouying" title="Maintenance">🚧</a> <a href="#infra-xpzouying" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#ideas-xpzouying" title="Ideas, Planning, & Feedback">🤔</a> <a href="#projectManagement-xpzouying" title="Project Management">📆</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/wendy57h"><img src="https://avatars.githubusercontent.com/u/168798147?v=4?s=100" width="100px;" alt="wendy57h"/><br /><sub><b>wendy57h</b></sub></a><br /><a href="https://github.com/better-world-ai/x-cli/commits?author=wendy57h" title="Code">💻</a> <a href="https://github.com/better-world-ai/x-cli/commits?author=wendy57h" title="Documentation">📖</a> <a href="#ideas-wendy57h" title="Ideas, Planning, & Feedback">🤔</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

## License

MIT — see [LICENSE](./LICENSE).
