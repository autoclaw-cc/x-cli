# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

[English](./README_EN.md) | 中文

你想在网页上反复做的事，一句话告诉 AI agent，它就能帮你做成 CLI 工具。生成的 CLI 让 agent 随时调用，直接驱动你真实的 Chrome 登录态，不走 API，不折腾 token。

仓库里收录了几个这样做出来的 CLI，既能装好就用，也作为参考案例，演示 AI agent + [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge) 是怎么从一句需求生成一个完整 CLI 的。

DEMO（一个 CLI 的诞生过程）：

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## 自己做一个新 CLI

仓库里几个 CLI 都是用 [`agent-cli-creator`](https://github.com/better-world-ai/agent-cli-creator) skill 让 AI agent 自动产出的。装好这个 skill，对你的 agent 说一句「帮我给 example.com 做个 CLI」就行。

完整的前置依赖（kimi-webbridge）、安装命令和使用步骤详见 [agent-cli-creator README](https://github.com/better-world-ai/agent-cli-creator)。

## 能做什么

5 个已经做好的场景：旅游规划、跨平台找房、高考志愿、AI 画图、搜索抓取——见 [Scenarios.md](./Scenarios.md)。

## 安装预编译二进制

> **前置**：CLI 需要驱动你本地的 Chrome，所以先装 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)（安装步骤见 [agent-cli-creator README → 前置依赖](https://github.com/better-world-ai/agent-cli-creator#前置依赖)）。

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

MIT，见 [LICENSE](./LICENSE)。
