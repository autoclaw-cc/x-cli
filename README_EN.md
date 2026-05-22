# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

English | [中文](./README.md)

Tell an AI agent in one sentence what you keep doing on a webpage, and it'll turn that into a CLI tool. The generated CLI can be called by your agent any time, driving your real Chrome login session directly — no API, no token juggling.

This repo collects several CLIs built exactly this way. They're installable and useful as-is, but they also serve as reference cases showing how AI agent + [kimi-webbridge](https://www.kimi.com/features/webbridge) turns a one-line request into a complete CLI.

DEMO (a CLI being born):

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## 5 ready-made scenarios

Every scenario needs [kimi-webbridge](https://www.kimi.com/features/webbridge) first (it drives your local Chrome). Install once, reuse everywhere.

### ✈️ Plan a full trip in one sentence

> "Plan a 5-day trip to Kyoto in June"

It's Friday night, you suddenly feel like getting out of town. Then you remember you'd have to open Ctrip to compare hotels, open Booking to check international rates, switch to a flight site to scan timings — and you put the idea down. Send that one sentence to the AI, and by the time you're done brushing your teeth, the flight times, hotel comparison, and attraction route are already lined up. Just follow it.

You can be more specific too: "May 1st, Shanghai to Chiang Mai round-trip, budget under $1500, pick two hotels rated 8.5+." The AI translates those constraints into the search itself.

**Uses**: [ctrip-cli](./ctrip-cli/) + [booking-cli](./booking-cli/) + [travel-planning](./skills/travel-planning/) skill

**Try it now**:
1. Download ctrip-cli and booking-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
2. `npx skills add better-world-ai/x-cli --skill travel-planning`
3. Open Claude and send the sentence above

---

### 🏠 Search rentals across countries in one shot

> "Find me a 1-bedroom in central London under £2000/month"

11 PM and you're still bouncing between 58.com, Anjuke, and Beike — refresh, load, retype filters, and half the listings turn out to be fake. Going abroad is worse: London means learning Rightmove, Madrid means Idealista, and every site's filter UI is different.

Tell the AI your budget, layout, and commute distance once. It queries five platforms in parallel, filters by your criteria, and gives you a single side-by-side list. Same prompt whether you're hunting domestic shares or long-term overseas rentals.

**Uses**: [58-cli](./58-cli/) + [anjuke-cli](./anjuke-cli/) + [apartments-cli](./apartments-cli/) + [rightmove-cli](./rightmove-cli/) + [idealista-cli](./idealista-cli/) + [rental-assistant](./skills/rental-assistant/) skill; helper `xiaohongshu-cli` for rental tips

**Try it now**:
1. Download the 5 rental CLIs from [Releases](https://github.com/better-world-ai/x-cli/releases)
2. `brew install xpzouying/agent-cli/xiaohongshu-cli`
3. `npx skills add better-world-ai/x-cli --skill rental-assistant`
4. Open Claude and send the sentence above

---

### 🎓 Gaokao admissions: all the info you need on one screen

> "I'm a Jiangsu test-taker with 580 points — which 211 schools can I get into?"

Scores are out. The application form is due in three days. Your provincial rank lands in some awkward bracket, prior-year admissions data is scattered across a dozen webpages and a few outdated PDFs, and the relatives' "experience-based advice" only makes it more confusing.

Ask the AI one question. It pulls official score lines, three years of admissions ranks, and corresponding majors, then ranks reach/target/safety schools based on your preferences. It won't decide for you, but everything you need to decide is on one screen.

**Uses**: [gaokao-cli](./gaokao-cli/) + [gaokao-assistant](./skills/gaokao-assistant/) skill

**Try it now**:
1. Download gaokao-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
2. `npx skills add better-world-ai/x-cli --skill gaokao-assistant`
3. Open Claude and send the sentence above

---

### 🎨 AI images without manually saving each one

> "Draw a shiba inu in a suit, standing in Times Square"

You need an image for a slide — open ChatGPT's web app, type the prompt, wait, right-click save, rename. Next image, same loop. By the tenth one you're sick of it, and the thought of batching thirty cover images for a layout is unbearable.

Describe what you want to the AI. It uses your already-logged-in Chrome to drive ChatGPT or Gemini, generates the images, and saves them to a local folder with consistent naming. No API key signup, no interruption to what you're doing — thirty images land on your desktop while you finish another paragraph.

**Uses**: [chatgpt-image-cli](./chatgpt-image-cli/) + [nanobanana-cli](./nanobanana-cli/)

**Try it now**:
1. Download chatgpt-image-cli or nanobanana-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
2. Open Claude and send the sentence above

---

### 🔍 Research a topic: search, read, summarize

> "Search for 'best local AI models 2025', fetch the body text of the top 10 results"

Want to understand an unfamiliar topic? The old way: Google it, click each result, read through, copy the key points, write up notes. A whole morning gone.

Hand that loop to the AI. It runs the search, follows each result, pulls the body text. You can ask it to synthesize a summary, or keep the raw articles to read yourself. For research scoping, tracking what's new in a field, or gathering material for an article, this is the front door.

**Uses**: [google-cli](./google-cli/) + [baidu-cli](./baidu-cli/)

**Try it now**:
1. Download google-cli or baidu-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
2. Open Claude and send the sentence above

---

## Install

> **Prerequisite**: install [kimi-webbridge](https://www.kimi.com/features/webbridge) first (drives your local Chrome, shared across all scenarios).

The 12 CLIs in this repo are distributed via GitHub releases. Grab the archive for your platform from the [Releases page](https://github.com/better-world-ai/x-cli/releases), extract it, and run.

`twitter-cli` and `xiaohongshu-cli` are distributed via Homebrew:

```bash
brew tap xpzouying/agent-cli
brew install twitter-cli xiaohongshu-cli
```

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

## Build your own scenario

The CLIs used in all 5 scenarios above were produced automatically by AI agents using the [`agent-cli-creator`](https://github.com/better-world-ai/agent-cli-creator) skill. Install the skill, then say "Build me a CLI for example.com" to your agent — that's it.

Full prerequisites, install commands, and walkthrough live in the [agent-cli-creator README](https://github.com/better-world-ai/agent-cli-creator/blob/main/README_EN.md).

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
