# x-cli
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

[English](./README_EN.md) | 中文

晚上 11 点，你还在 58、安居客、贝壳之间来回切，刷新、加载、重新填条件，看到的房源还有一半是假的。换个国家更头大，去伦敦要会 Rightmove，去马德里要会 Idealista，每个站点的筛选器都长得不一样。

我没去应用商店找工具。我对 agent 说了一句话：

> 帮我给 58 同城做个 CLI，我要能搜租房。

然后我去干了点别的。回来的时候，`58-cli` 能跑了。后来安居客、Rightmove、Idealista，同一句话又重复了几遍。

这件事靠两个东西撑着：

- **[agent-cli-creator](https://github.com/better-world-ai/agent-cli-creator)** —— 一个 skill，教 agent 怎么把一个网站变成 CLI。它管「怎么造」。
- **[kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)** —— 一个浏览器插件加一个本地 skill，让 agent 能开你桌上那台 Chrome。它管「造出来的东西靠什么跑」。

为什么非得绕浏览器一圈？因为 58 同城没有开放 API，安居客没有，携程也没有。这些站的数据不在任何一份接口文档里，它只存在于你已经打开、已经登录好的那个标签页里。webbridge 做的事，就是把那个标签页交到 agent 手上——不用申请 key，不用管 token 什么时候过期。

这个仓库是成品展示区。里面 15 个 CLI，都是这么来的。

## 那句话之后，agent 干了什么

它先反问了我两个问题：用什么语言，先做哪几个功能。我说 Go，先做搜索和详情两个就够。

然后它没有立刻写代码。它开了我的 Chrome，真的去 58 同城搜了一次房，盯着网络请求把接口扭了出来，又在浏览器里把这个接口原地跑通一次，确认真能拿到数据。**整个流程里最关键的就是这一步**——站点不会告诉你它的接口长什么样，只能自己去看一眼。

跑通之后才动手。搭脚手架，先写不改数据的读命令，写完一个跑一遍验证，再写下一个。

最后我拿到的东西：

```bash
58-cli search --city sh --keyword 张江 --max-price 5000 --limit 3
```

```json
{"ok": true, "data": {"listings": [{"title": "...", "rent_monthly": 4800, "layout": "2室1厅", "area_sqm": 68}]}}
```

一个 CLI 的完整诞生过程，录在这里：

https://github.com/user-attachments/assets/c1d04187-972a-4b8a-b243-df085281fc77

## 同一句话，重复了 14 遍

| 场景 | 用到的 CLI | 完整脚本 |
|------|-----------|---------|
| 跨国租房 | [58-cli](./58-cli/)、[anjuke-cli](./anjuke-cli/)、[apartments-cli](./apartments-cli/)、[rightmove-cli](./rightmove-cli/)、[idealista-cli](./idealista-cli/) + [rental-assistant](./skills/rental-assistant/) skill | [find-shanghai-rental](./recipes/find-shanghai-rental.md) |
| 行程规划 | [ctrip-cli](./ctrip-cli/)、[booking-cli](./booking-cli/) + [travel-planning](./skills/travel-planning/) skill | [plan-kyoto-trip](./recipes/plan-kyoto-trip.md) |
| 高考志愿 | [gaokao-cli](./gaokao-cli/) + [gaokao-assistant](./skills/gaokao-assistant/) skill | [gaokao-jiangsu-211](./recipes/gaokao-jiangsu-211.md) |
| AI 批量出图 | [chatgpt-image-cli](./chatgpt-image-cli/)、[nanobanana-cli](./nanobanana-cli/) | [batch-image-shiba-inu](./recipes/batch-image-shiba-inu.md) |
| 主题调研 | [google-cli](./google-cli/)、[baidu-cli](./baidu-cli/) | [research-local-ai-models](./recipes/research-local-ai-models.md) |
| 文献检索 | [scholar-cli](./scholar-cli/) + [paper-research](./skills/paper-research/) skill | [review-rag-literature](./recipes/review-rag-literature.md) |
| 找工作 | [boss-cli](./boss-cli/) | 暂无 |
| 免费配图 | [unsplash-cli](./unsplash-cli/) | 暂无 |

另有 `twitter-cli` 和 `xiaohongshu-cli` 通过 Homebrew 分发，不在本仓库目录里：

```bash
brew tap xpzouying/agent-cli
brew install twitter-cli xiaohongshu-cli
```

## 你要装什么，看你想干嘛

### 只想用现成的 CLI

**不需要 agent-cli-creator。** 两步就够：

**第一步，装 kimi-webbridge。** 它分两部分，装一次，之后所有 CLI 共用：

1. 浏览器插件，agent 控制浏览器的入口。装好之后所有点击、输入、读取都通过它转发，你登录过的 Chrome 会话自动被复用。
   - 中文：<https://www.kimi.com/zh-cn/features/webbridge>
   - English：<https://www.kimi.com/features/webbridge>
2. 本地 skill，让 agent 知道怎么用上面那个插件：

   ```bash
   curl -fsSL https://cdn.kimi.com/webbridge/install.sh | bash
   ```

**第二步，下 CLI。** 去 [Releases 页面](https://github.com/better-world-ai/x-cli/releases) 下对应平台的归档，解压即可用。

上面表格里带 skill 的场景，再补一句就行，比如：

```bash
npx skills add better-world-ai/x-cli --skill rental-assistant
```

然后打开你的 agent，直接说人话：「我在上海张江找两室一厅，月租 5000 以内」。

### 想自己造一个

1. 先装 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)（同上两部分，已经装过就跳过）
2. 装 [agent-cli-creator](https://github.com/better-world-ai/agent-cli-creator)：

   ```bash
   npx skills add better-world-ai/agent-cli-creator
   ```

3. 在 Chrome 里登录目标网站，然后对 agent 说：「帮我给 example.com 做个 CLI，我要能拉首页信息流，并且能发评论。」

agent 会先问你用什么语言、先做哪几个功能，然后自己去探站点、搭脚手架、写命令，关键节点停下来跟你确认。没有 Node.js 的手动安装方式、以及这个 skill 内部怎么工作，见 [agent-cli-creator README](https://github.com/better-world-ai/agent-cli-creator)。

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

## 轮到你了

代码早就不是瓶颈了。瓶颈是你愿不愿意把那句话说出口。

```bash
npx skills add better-world-ai/agent-cli-creator
```

然后把空填上，发给你的 agent：

> 帮我给 \_\_\_\_\_\_ 做个 CLI，我要能 \_\_\_\_\_\_。

仓库里这 14 个，当初都是这么开始的。

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
