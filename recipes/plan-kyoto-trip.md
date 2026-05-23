# 一句话规划京都行程

[English](./plan-kyoto-trip_EN.md) | 中文

## 场景

周末突然想出门走走，但不想花整天比酒店、机票、景点。让 AI 把这些活一次干完，给你一份能照着执行的清单。

## 前置

- [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)（驱动本地 Chrome）
- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 ctrip-cli、booking-cli
- 装 skill：`npx skills add better-world-ai/x-cli --skill travel-planning`

## 发给 Claude

```
帮我规划 6 月份去京都的 5 天行程，预算控制在 1.5 万以内。
偏好和服体验、精品咖啡店和深度寺庙游览，不爱购物。
酒店挑评分 8.5+、靠近地铁站的两间，对比国内（携程）和国际（Booking）的报价取低。
机票上海出发，避开红眼航班。
```

## 你会拿到

- 推荐机票（航司、时段、价格区间、转机情况）
- 两到三间候选酒店（评分、位置、携程 vs Booking 价格对比）
- 5 天每日行程（景点、动线、预计耗时、营业时间）
- 总预算分解（机票 + 住宿 + 景点门票 + 餐饮估算）

AI 会在关键节点停下来跟你确认是否调整，不是一次跑完。

## 用到

[ctrip-cli](../ctrip-cli/) + [booking-cli](../booking-cli/) + [travel-planning](../skills/travel-planning/) skill

## 改一改

把京都换成任何城市、把日期/预算换成你的实际情况、把偏好换成你自己的。带娃可以加亲子景点权重，蜜月可以加情侣餐厅权重，直接说就行。
