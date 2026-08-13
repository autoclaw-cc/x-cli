# 用一句话完成一轮文献调研

[English](./research-academic-papers_EN.md) | 中文

## 场景

准备开题、写综述或快速摸清一个研究方向时，不想在多个数据库之间反复搜索、补元数据、下载 PDF，再手工整理引用。让 AI 把检索、筛选、补全、下载和 BibTeX 导出串成一条可检查的调研链路。

## 前置

- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 scholar-cli
- 装 skill：`npx skills add better-world-ai/x-cli --skill paper-research`
- 只有使用 Google Scholar、知网或 Web of Science 时才需要 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)；知网和 WoS 还需要在浏览器中完成登录

## 发给 Claude

```
帮我调研“检索增强生成（RAG）的评测方法”，把工作区放在 ./research/rag-evaluation。

先用 scholar-cli search-en 搜索英文文献，限定 arxiv,semantic 两个源，取 15 篇并存入工作区。按相关性选出 5 篇有 DOI 的核心论文，逐篇用 detail --doi 补全作者、摘要、年份和引用信息。对有合法开放下载渠道的论文执行 download，把 PDF 存到 ./research/rag-evaluation/papers；不要启用 --scihub。最后用 export 从工作区导出 ./research/rag-evaluation/refs.bib。

返回一张核心论文表，说明每篇为什么值得读、PDF 是否下载成功，并列出失败的源或缺失的 DOI。不要把未确认的信息补写成事实。
```

## 你会拿到

- 一个自动去重、可继续追加搜索的文献工作区
- 5 篇核心论文的完整元数据和阅读理由
- 能合法开放获取的 PDF，以及未下载项目的明确原因
- 可直接导入 Zotero、LaTeX 等工具的 `refs.bib`
- 各数据源失败、限流或缺少 DOI 时的清单，而不是被静默丢掉的结果

## 用到

[scholar-cli](../scholar-cli/) + [paper-research](../skills/paper-research/) skill

## 改一改

把 RAG 换成你的研究主题，并按学科调整 `--sources`。中文文献可加 `search-cnki`（需先在浏览器登录并通过可能出现的滑块验证）；需要 Web of Science 时，先运行 `login-status --platform wos` 确认机构登录状态，再用 `search-wos` 检索。可以调整核心论文数量、工作区和导出路径，但不要把 `--scihub` 写进默认流程。
