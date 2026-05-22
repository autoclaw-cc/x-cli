# 找上海张江两室一厅

## 场景

工作搬迁需要找房，预算和通勤都有约束。一次问完几个国内平台，避免一个个站点切换填条件，顺便看看小红书上有没有避坑提醒。

## 前置

- [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)
- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 58-cli、anjuke-cli
- 装小红书 CLI（查攻略和避坑）：`brew install xpzouying/agent-cli/xiaohongshu-cli`
- 装 skill：`npx skills add better-world-ai/x-cli --skill rental-assistant`

## 发给 Claude

```
帮我在上海张江找两室一厅，月租 5000 以内。
通勤目标是张江高科地铁站 30 分钟内，电梯房、有独立卫浴的优先。
58 和安居客都查一下做交叉对照，再去小红书看看这一片有没有特别的避坑提醒。
```

## 你会拿到

- 跨平台合并去重的房源列表（标题、价格、面积、户型、距离、原链接）
- 价格分布概况，按小区或区域分组
- 小红书上跟这片相关的近期帖子摘要（避坑、转租注意事项、中介评价）
- 看到的明显异常（同一房源在两个平台价格差很大、明显的钓鱼帖等）

## 用到

[58-cli](../58-cli/) + [anjuke-cli](../anjuke-cli/) + `xiaohongshu-cli` + [rental-assistant](../skills/rental-assistant/) skill

## 改一改

- **换城市**：把张江换成成都春熙路、深圳南山等等
- **换户型**：合租单间、整租三室、Loft 都可以指定
- **出国版**：换成「伦敦 Zone 2 一居室，月租 2000 镑以内」会自动切到 Rightmove；换「马德里中心区一居室」会切 Idealista；换「波士顿 Cambridge 区一居室」会切 Apartments.com
- **加美区**：先装好 apartments-cli、rightmove-cli、idealista-cli，描述里说出哪个国家/城市即可
