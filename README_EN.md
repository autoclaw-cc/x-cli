# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

English | [中文](./README.md)

11 PM, and you're still bouncing between Zillow, Apartments.com and half a dozen local listing sites. Refresh, wait, re-enter the filters, and half of what comes back is stale anyway. Try another country and it gets worse — London means Rightmove, Madrid means Idealista, and every site's filter panel is shaped differently.

I didn't go looking for a tool. I said one sentence to my agent:

> Build me a CLI for 58.com. I want to search rentals.

Then I went and did something else. By the time I came back, `58-cli` worked. Anjuke, Rightmove, Idealista came later — the same sentence, a few more times.

Two things make this work:

- **[agent-cli-creator](https://github.com/better-world-ai/agent-cli-creator)** — a skill that teaches your agent how to turn a website into a CLI. It handles *how to build*.
- **[kimi-webbridge](https://www.kimi.com/features/webbridge)** — a browser extension plus a local skill that lets the agent drive the Chrome already sitting on your desk. It handles *what the result runs on*.

Why route through a browser at all? Because 58.com has no public API. Neither does Anjuke. Neither does Ctrip. That data doesn't live in any API doc — it only exists inside the tab you already have open and already logged into. What webbridge does is hand that tab to your agent. No key to apply for, no token to keep alive.

This repo is the showcase of what came out. All 15 CLIs in it were built this way.

## What the agent did after that sentence

It asked me two questions first: which language, and which features to start with. I said Go, and just search plus detail for now.

Then it didn't write any code. It opened my Chrome, actually searched 58.com for an apartment, watched the network requests until it had the endpoint, and then ran that endpoint right there in the browser to confirm real data came back. **That step is the one that matters most** — a site won't tell you what its endpoints look like, so you have to go look.

Only after that did it start building. Scaffold, then read-only commands first, verifying each one before moving to the next.

What I ended up with:

```bash
58-cli search --city sh --keyword 张江 --max-price 5000 --limit 3
```

```json
{"ok": true, "data": {"listings": [{"title": "...", "rent_monthly": 4800, "layout": "2室1厅", "area_sqm": 68}]}}
```

Here's a full CLI being born, start to finish:

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## The same sentence, 14 more times

| Scenario | CLIs used | Full recipe |
|----------|-----------|-------------|
| Rentals across countries | [58-cli](./58-cli/), [anjuke-cli](./anjuke-cli/), [apartments-cli](./apartments-cli/), [rightmove-cli](./rightmove-cli/), [idealista-cli](./idealista-cli/) + [rental-assistant](./skills/rental-assistant/) skill | [find-shanghai-rental](./recipes/find-shanghai-rental_EN.md) |
| Trip planning | [ctrip-cli](./ctrip-cli/), [booking-cli](./booking-cli/) + [travel-planning](./skills/travel-planning/) skill | [plan-kyoto-trip](./recipes/plan-kyoto-trip_EN.md) |
| Gaokao applications | [gaokao-cli](./gaokao-cli/) + [gaokao-assistant](./skills/gaokao-assistant/) skill | [gaokao-jiangsu-211](./recipes/gaokao-jiangsu-211_EN.md) |
| Batch image generation | [chatgpt-image-cli](./chatgpt-image-cli/), [nanobanana-cli](./nanobanana-cli/) | [batch-image-shiba-inu](./recipes/batch-image-shiba-inu_EN.md) |
| Topic research | [google-cli](./google-cli/), [baidu-cli](./baidu-cli/) | [research-local-ai-models](./recipes/research-local-ai-models_EN.md) |
| Academic literature | [scholar-cli](./scholar-cli/) + [paper-research](./skills/paper-research/) skill | [review-rag-literature](./recipes/review-rag-literature_EN.md) |
| Job hunting | [boss-cli](./boss-cli/) | — |
| Free stock photos | [unsplash-cli](./unsplash-cli/) | — |

`twitter-cli` and `xiaohongshu-cli` ship via Homebrew and don't live in this repo:

```bash
brew tap xpzouying/agent-cli
brew install twitter-cli xiaohongshu-cli
```

## What to install depends on what you want

### Just want to use the existing CLIs

**You don't need agent-cli-creator.** Two steps:

**Step 1 — install kimi-webbridge.** It comes in two parts, installed once and shared by every CLI here:

1. The browser extension — your agent's way into the browser. Once it's in, every click, keystroke and read goes through it, and the Chrome sessions you're already logged into get reused automatically.
   - English: <https://www.kimi.com/features/webbridge>
   - 中文: <https://www.kimi.com/zh-cn/features/webbridge>
2. The local skill, which teaches your agent how to use that extension:

   ```bash
   curl -fsSL https://cdn.kimi.com/webbridge/install.sh | bash
   ```

**Step 2 — grab a CLI.** Download the archive for your platform from the [Releases page](https://github.com/better-world-ai/x-cli/releases) and extract it.

For the scenarios above that come with a skill, add one more line, e.g.:

```bash
npx skills add better-world-ai/x-cli --skill rental-assistant
```

Then open your agent and just talk: "Find me a two-bedroom in Zhangjiang, Shanghai, under 5000 a month."

### Want to build your own

1. Install [kimi-webbridge](https://www.kimi.com/features/webbridge) (both parts above — skip if you already have it)
2. Install [agent-cli-creator](https://github.com/better-world-ai/agent-cli-creator):

   ```bash
   npx skills add better-world-ai/agent-cli-creator
   ```

3. Log into the target site in Chrome, then tell your agent: "Build me a CLI for example.com. I want to pull the home feed and post comments."

The agent will ask which language and which features to start with, then go explore the site, scaffold the project and implement the commands, stopping to check with you at the decision points. For manual installation without Node.js, and for how the skill works internally, see the [agent-cli-creator README](https://github.com/better-world-ai/agent-cli-creator).

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

## Your turn

Code stopped being the bottleneck a while ago. The bottleneck is whether you say the sentence out loud.

```bash
npx skills add better-world-ai/agent-cli-creator
```

Then fill in the blanks and send it to your agent:

> Build me a CLI for \_\_\_\_\_\_. I want to \_\_\_\_\_\_.

All 15 in this repo started exactly there.

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
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/RachelXiaolan"><img src="https://avatars.githubusercontent.com/u/236927962?v=4?s=100" width="100px;" alt="Rachel Lu"/><br /><sub><b>Rachel Lu</b></sub></a><br /><a href="https://github.com/better-world-ai/x-cli/commits?author=RachelXiaolan" title="Code">💻</a> <a href="https://github.com/better-world-ai/x-cli/pulls?q=is%3Apr+reviewed-by%3ARachelXiaolan" title="Reviewed Pull Requests">👀</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

## License

MIT — see [LICENSE](./LICENSE).
