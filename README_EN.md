# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

English | [中文](./README.md)

Tell an AI agent in one sentence what you keep doing on a webpage, and it'll turn that into a CLI tool. The generated CLI can be called by your agent any time, driving your real Chrome login session directly — no API, no token juggling.

This repo collects several CLIs built exactly this way. They're installable and useful as-is, but they also serve as reference cases showing how AI agent + [kimi-webbridge](https://www.kimi.com/features/webbridge) turns a one-line request into a complete CLI.

DEMO (a CLI being born):

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## Build your own CLI

All the CLIs in this repo were produced automatically by AI agents using the [`agent-cli-creator`](https://github.com/better-world-ai/agent-cli-creator) skill. Install the skill, then say "Build me a CLI for example.com" to your agent — that's it.

Full prerequisites (kimi-webbridge), install commands, and walkthrough live in the [agent-cli-creator README](https://github.com/better-world-ai/agent-cli-creator/blob/main/README_EN.md).

## What you can do

Five ready-made scenarios: travel planning, cross-platform rentals, gaokao admissions, AI image generation, and search/scrape — see [Scenarios_EN.md](./Scenarios_EN.md).

## Install prebuilt binaries

> **Prerequisite**: the CLIs drive your local Chrome, so first install [kimi-webbridge](https://www.kimi.com/features/webbridge) (see [agent-cli-creator README → Prerequisites](https://github.com/better-world-ai/agent-cli-creator/blob/main/README_EN.md#prerequisites)).

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
