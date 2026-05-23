# Recipes

[English](./README_EN.md) | 中文

完整的「复制粘贴就能用」的脚本集合。每一份对应 [主 README](../README.md#现成的-5-个场景) 里的一个场景，给出可直接发给 Claude 的提示词、前置步骤和预期产出。

## 索引

| Recipe | 场景 | 用到 |
|---|---|---|
| [plan-kyoto-trip](./plan-kyoto-trip.md) | 规划完整旅行行程 | ctrip-cli + booking-cli + travel-planning |
| [find-shanghai-rental](./find-shanghai-rental.md) | 跨平台找房（含海外） | 58-cli + anjuke-cli + apartments-cli + rightmove-cli + idealista-cli + xiaohongshu-cli + rental-assistant |
| [gaokao-jiangsu-211](./gaokao-jiangsu-211.md) | 高考志愿冲稳保 | gaokao-cli + gaokao-assistant |
| [batch-image-shiba-inu](./batch-image-shiba-inu.md) | 批量 AI 出图 | chatgpt-image-cli 或 nanobanana-cli |
| [research-local-ai-models](./research-local-ai-models.md) | 调研话题 + 抓正文 + 整理 | google-cli 或 baidu-cli |

## 怎么贡献新 recipe

按现有 recipe 的结构写一份 markdown 文件，提交 PR。建议每份包含：

- **场景**：什么样的真实需求会用到
- **前置**：装哪几个 CLI、需要登录哪些站点
- **发给 Claude**：可以直接复制的提示词（用代码块）
- **你会拿到**：描述产出结构
- **用到**：CLI 和 skill 链接
- **改一改**：怎么调整提示词适应你自己的场景
