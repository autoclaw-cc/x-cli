# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

[English](./README_EN.md) | 中文

你想在网页上反复做的事，一句话告诉 AI agent，它就能帮你做成 CLI 工具。生成的 CLI 让 agent 随时调用，直接驱动你真实的 Chrome 登录态，不走 API，不折腾 token。

仓库里收录了几个这样做出来的 CLI，既能装好就用，也作为参考案例，演示 AI agent + [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge) 是怎么从一句需求生成一个完整 CLI 的。

DEMO（一个 CLI 的诞生过程）：

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## 现成的 5 个场景

所有场景都需要先装 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)（驱动你本地 Chrome），只装一次。

### ✈️ 一句话规划完整行程

> "帮我规划 6 月份去京都的 5 天行程"

周五晚上突然想出门走走，一想到要打开携程比酒店、再开 Booking 看一遍海外报价、再切到机票网站排时段，又把这个念头放下了。把这句话发给 AI，等你刷完牙回来，机票时段、酒店对比、景点动线已经排好，直接照着走。

也可以让它做更具体的事，"5 月 1 日上海到清迈往返，预算 1 万以内，挑两间评分 8.5+ 的酒店"，约束条件 AI 会自己落到搜索里。

**用到**：[ctrip-cli](./ctrip-cli/) + [booking-cli](./booking-cli/) + [travel-planning](./skills/travel-planning/) skill

**现在就试**：
1. 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下 ctrip-cli、booking-cli
2. `npx skills add better-world-ai/x-cli --skill travel-planning`
3. 打开 Claude，发上面那句话

完整脚本：[recipes/plan-kyoto-trip.md](./recipes/plan-kyoto-trip.md)

---

### 🏠 一次找完几个国家的房

> "我在上海张江找两室一厅，月租 5000 以内"

晚上 11 点还在 58、安居客、贝壳之间来回切，刷新、加载、重新填条件，看到的房源还有一半是假的。换个国家更头大，去伦敦要会 Rightmove，去马德里要会 Idealista，每个站点的筛选器都长得不一样。

把你的预算、户型、通勤距离一次说清楚，AI 把五个平台同时跑一遍，按你的条件过滤好，给你一份对照清单。从国内合租到海外长租，话术不变。

**用到**：[58-cli](./58-cli/) + [anjuke-cli](./anjuke-cli/) + [apartments-cli](./apartments-cli/) + [rightmove-cli](./rightmove-cli/) + [idealista-cli](./idealista-cli/) + [rental-assistant](./skills/rental-assistant/) skill；辅助工具 `xiaohongshu-cli` 可查租房攻略

**现在就试**：
1. 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下五个 rental CLI
2. `brew install xpzouying/agent-cli/xiaohongshu-cli`
3. `npx skills add better-world-ai/x-cli --skill rental-assistant`
4. 打开 Claude，发上面那句话

完整脚本：[recipes/find-shanghai-rental.md](./recipes/find-shanghai-rental.md)

---

### 🎓 高考志愿，把决定要的信息摆齐

> "我是江苏考生，580 分能上哪些 211"

分数出来了，志愿表三天后就要交。位次卡在某个尴尬区间，往年录取数据散在十几个网页和几份过期 PDF 里，家里亲戚的"经验之谈"听完更迷茫。

问 AI 一句，它把官方分数线、近三年录取位次、对应专业拉齐，结合你的偏好排出冲、稳、保三档。它不替你做选择，但你做选择需要的所有数据，能在一个清单里看完。

**用到**：[gaokao-cli](./gaokao-cli/) + [gaokao-assistant](./skills/gaokao-assistant/) skill

**现在就试**：
1. 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下 gaokao-cli
2. `npx skills add better-world-ai/x-cli --skill gaokao-assistant`
3. 打开 Claude，发上面那句话

完整脚本：[recipes/gaokao-jiangsu-211.md](./recipes/gaokao-jiangsu-211.md)

---

### 🎨 让 AI 画图，不用一张一张右键保存

> "画一只穿西装的柴犬，站在 Times Square"

做 PPT 临时缺一张图，开 ChatGPT 网页版输入提示词、等出图、右键保存、改文件名，下一张再来一遍。做到第十张就够烦了，更别说要批量产三十张排版用的封面图。

把你要的图描述给 AI，它用你已经登录的 Chrome 直接调 ChatGPT 或 Gemini 出图，按命名规则落到本地文件夹。没有 API key 注册，不打断你正在做的事，三十张图等你写完一段文档就在桌面上。

**用到**：[chatgpt-image-cli](./chatgpt-image-cli/) + [nanobanana-cli](./nanobanana-cli/)

**现在就试**：
1. 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下 chatgpt-image-cli 或 nanobanana-cli
2. 打开 Claude，发上面那句话

完整脚本：[recipes/batch-image-shiba-inu.md](./recipes/batch-image-shiba-inu.md)

---

### 🔍 一个话题搜完、读完、整理完

> "搜一下 2025 年值得用的本地 AI 模型，把前 10 篇正文都拿回来"

想了解一个不熟悉的话题，老办法是 Google 搜一下、每个结果点进去读完、复制重点、整理成笔记，一上午没了。

这段流程让 AI 替你跑，它跑搜索、顺着结果抓正文，你可以直接让它综合成一份摘要，也可以保留原文自己看。研究选题、看一个领域的新进展、找资料写文章，先用它把信息汇总到一处。

**用到**：[google-cli](./google-cli/) + [baidu-cli](./baidu-cli/)

**现在就试**：
1. 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下 google-cli 或 baidu-cli
2. 打开 Claude，发上面那句话

完整脚本：[recipes/research-local-ai-models.md](./recipes/research-local-ai-models.md)

---

## 网页 AI 账号 CLI

- [chatgpt-cli](./chatgpt-cli/)：普通问答、网页搜索、Deep Research、引用提取，以及图片生成并落盘。
- [notebooklm-cli](./notebooklm-cli/)：自有笔记本、来源、问答、笔记、Fast/Deep Research，以及九类 Studio 产物的生成和管理。
- [deepseek-cli](./deepseek-cli/)：普通问答、DeepThink、联网搜索、文件与图片输入。

三者都只通过 Kimi WebBridge 使用隔离 Chrome 的既有登录态，不导出 Cookie、token 或账号身份。详细命令和安全边界见各自 README。

## 安装

> **前置**：先装 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)（驱动本地 Chrome，所有场景共用）。

本仓库的 15 个 CLI 通过 GitHub releases 分发，去 [Releases 页面](https://github.com/better-world-ai/x-cli/releases) 下载对应平台归档，解压即可用。

另有 `twitter-cli` 和 `xiaohongshu-cli` 通过 Homebrew 分发：

```bash
brew tap xpzouying/agent-cli
brew install twitter-cli xiaohongshu-cli
```

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

## 自己做一个新场景

上面 5 个场景里用到的 CLI，都是用 [`agent-cli-creator`](https://github.com/better-world-ai/agent-cli-creator) skill 让 AI agent 自动产出的。装好这个 skill，对你的 agent 说一句「帮我给 example.com 做个 CLI」就行。

完整的前置依赖、安装命令和使用步骤详见 [agent-cli-creator README](https://github.com/better-world-ai/agent-cli-creator)。

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
