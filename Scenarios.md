# 能做什么

[English](./Scenarios_EN.md) | 中文

下面 5 个场景都是已经做好的——对 AI 说一句中文需求就能跑。CLI 是底下的工具，AI 替你串起来。

## ✈️ 一句话规划完整行程

> "帮我规划 6 月份去京都的 5 天行程"

让 AI 同时跑携程和 Booking——对比国内外的酒店报价、查机票、列出必玩景点，最后给你一份可执行的行程清单。

用到：[ctrip-cli](./ctrip-cli/) · [booking-cli](./booking-cli/) · [travel-planning](./skills/travel-planning/)

## 🏠 一次查多个网站找房

> "我在上海张江找两室一厅，月租 5000 以内"

AI 同时查 58同城、安居客、Apartments.com、Rightmove、Idealista——价格、户型、距离都按你说的过滤好，给你一份跨平台对比的房源列表。不管你找的是国内合租还是海外长租。

用到：[58-cli](./58-cli/) · [anjuke-cli](./anjuke-cli/) · [apartments-cli](./apartments-cli/) · [rightmove-cli](./rightmove-cli/) · [idealista-cli](./idealista-cli/) · [rental-assistant](./skills/rental-assistant/)

## 🎓 高考志愿辅助

> "我是江苏考生，580 分能上哪些 211"

AI 查官方分数线和往年录取数据，结合你的位次和偏好，给你一份冲 / 稳 / 保的志愿建议。

用到：[gaokao-cli](./gaokao-cli/) · [gaokao-assistant](./skills/gaokao-assistant/)

## 🎨 用 AI 画图

> "画一只穿西装的柴犬，站在 Times Square"

让 AI 通过 ChatGPT 网页版或 Google Gemini 出图，自动保存到本地。不用复制粘贴，不用注册 API key，直接用你已经登录的 Chrome。

用到：[chatgpt-image-cli](./chatgpt-image-cli/) · [nanobanana-cli](./nanobanana-cli/)

## 🔍 搜索 + 抓取网页

> "搜索关键词 X 的前 10 个 Google 结果，把每个页面的正文都抓回来"

让 AI 跑 Google 或百度搜索，然后顺着结果点进去把详情都抓回来。是搭建更复杂工作流的底层积木。

用到：[google-cli](./google-cli/) · [baidu-cli](./baidu-cli/)

---

想做新的场景？用 [`agent-cli-creator`](./skills/agent-cli-creator/) skill——一句中文需求，AI 帮你做出新的 CLI。详见主 README 的「自己做一个新 CLI」。
