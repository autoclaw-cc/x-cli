# 一句话做完一轮文献综述

[English](./review-rag-literature_EN.md) | 中文

## 场景

你要摸清一个不熟悉的研究方向。老办法是 Google Scholar 搜一遍、arXiv 再搜一遍、把重复的手动去掉、一篇篇点进去看摘要、有 PDF 的存到本地、最后再手敲一遍 BibTeX。一上午就没了，而且下次换个题目还得重来。

这条链路让 agent 替你跑：搜索、去重、看细节、下 PDF、导出参考文献，一次说清楚就行。

**这个场景不需要任何账号。** arXiv、Semantic Scholar、CrossRef、OpenAlex 都是公开 API，不用注册、不用 API key。

## 前置

- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 scholar-cli
- 装 skill：`npx skills add better-world-ai/x-cli --skill paper-research`
- 可选：`export SCHOLAR_CLI_EMAIL=你的邮箱` —— CrossRef、OpenAlex、Unpaywall 都属于「礼貌池」，留个联系邮箱能拿到更稳的配额。**Unpaywall 会直接拒绝默认的占位邮箱**，不设这个变量就少一条 PDF 下载渠道。

只搜公开源的话，**连 kimi-webbridge 都不需要**。只有想加上 Google Scholar 时才要装它（见文末）。

## 发给 Claude

```
帮我做一轮 RAG 评测方向的文献调研。

搜 arXiv、Semantic Scholar 和 CrossRef，关键词 "retrieval augmented
generation evaluation"，每个源取 8 篇，结果存到 ./rag-review 这个工作区里，
重复的自动合并。

搜完先给我一份清单：标题、年份、来源、有没有 DOI。
我挑几篇之后，你再去补完整信息，能下 PDF 的下到 ./rag-review/pdfs，
最后导出一份 BibTeX 给我。
```

## 你会拿到

**第一步，跨源搜索并自动去重。**三个源各返回 8 篇，去重后 16 篇进入工作区：

```
papers_added: 16
sources: [{'name': 'arxiv', 'count': 8},
          {'name': 'semantic', 'count': 0, 'error': '...rate limited (HTTP 429) after 4 attempts'},
          {'name': 'crossref', 'count': 8}]
```

注意中间那行：**某个源被限流时它会明说，另外两个照常返回**。你拿到的是 16 篇加一句诚实的说明，而不是一个假装成功的空结果。

**第二步，补全单篇信息。**

```
title    : MODE-RAG: Manifold Outlier Diagnosis and Energy-based Retrie...
year     : 2026 | venue: Proceedings of the 2nd Workshop on Multi...
authors  : Zehang Wei, JiaXin Dai, Jiamin Yan
sources  : ['crossref', 'semantic']
```

arXiv 上的论文也一样能查，它会自动走 arXiv 的接口：

```
title  : Retrieval-Augmented Generation for Knowledge-Intensive NLP T...
sources: ['arxiv'] | year: 2020
```

**第三步，下载开放获取的 PDF。**

```
arxiv | 885323 bytes | Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks.pdf
```

**第四步，导出 BibTeX。**工作区里 16 篇一次导出，可以直接丢进 LaTeX：

```bibtex
@misc{qi2025arrag,
  title = {AR-RAG: Autoregressive Retrieval Augmentation for Image Generation},
  author = {Jingyuan Qi and Zhiyang Xu and Qifan Wang and Lifu Huang},
  year = {2025},
  url = {https://arxiv.org/pdf/2506.06962v3},
  eprint = {2506.06962},
  archiveprefix = {arXiv},
}
```

工作区是累加的。同一个目录下反复搜不同关键词，重复的论文会自动合并，最后一次性导出。

## 用到

[scholar-cli](../scholar-cli/) + [paper-research](../skills/paper-research/) skill

## 加上 Google Scholar（可选）

Google Scholar 的独门优势是**引用数**，公开 API 给不了：

```
· Ragas: Automated evaluation of retrieval augmented g...
    S Es, J James, LE Anke | Proceedings of the 18th … | 2024 | cites: 2041
```

想用它，多两个前提：

1. 装 [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)（它驱动你本地的 Chrome）
2. Google Scholar **可能弹验证码**。弹了就自己在浏览器里点一下，再让 agent 重试

在提示词里加一句「再用 Google Scholar 搜一遍，我想看引用数」就行。

## 需要机构权限的两个源

`scholar-cli` 还支持 CNKI（知网）和 Web of Science，但它们要登录：知网需要你已登录并过了滑块验证，WoS 需要机构 VPN 或 SSO。有条件的话在提示词里直接说「用知网也搜一遍」。**没有的话不影响上面任何一步。**

## 改一改

把关键词换成你自己的方向就行。几个能直接说的变化：

- **只要最新的**：「只看 2024 年以后的」
- **只要能下到全文的**：「跳过下不到 PDF 的」
- **分批调研**：换关键词重跑，指同一个工作区，去重是自动的
- **中文文献**：如果你有知网权限，加一句「中文的用知网搜」
