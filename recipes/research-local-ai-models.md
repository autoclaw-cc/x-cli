# 调研「2025 本地 AI 模型」话题

[English](./research-local-ai-models_EN.md) | 中文

## 场景

要快速摸清一个不熟悉的话题——比如想看看 2025 年的本地 AI 模型生态怎么样了。手动一篇篇 Google 读过来太慢，让 AI 跑搜索、抓正文、整理结论。

## 前置

- [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)
- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 google-cli 或 baidu-cli

## 发给 Claude

```
帮我搜 "best local AI models 2025"，把 Google 前 10 个结果的正文都抓回来。
然后做两件事：
1. 用 200 字左右总结这些文章的共识观点和分歧点
2. 列一张表，整理这些文章提到的所有模型，按推荐次数排序

中文输出，原始链接附在最后。
```

## 你会拿到

- 一段精炼的综合摘要（共识 + 分歧）
- 一张模型对比表（名称、推荐它的文章数、主要优势、典型用例）
- 原始结果链接列表，方便点进去看细节
- 如果某些页面抓取失败（付费墙、403），AI 会标注跳过

## 用到

[google-cli](../google-cli/) 或 [baidu-cli](../baidu-cli/)

## 改一改

- **中文话题**：用 baidu-cli 替代 google-cli
- **调整数量**：前 5 个还是前 20 个都行
- **加约束**：只看近 30 天的、排除特定网站、按时间排序
- **换输出形式**：摘要、对比表、SWOT、问答 FAQ、PPT 大纲都能直接说
