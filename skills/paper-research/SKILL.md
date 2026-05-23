---
name: paper-research
description: Use when the user wants to search academic papers, find literature, download PDFs, or export citations. Invoke when user mentions "论文", "文献", "paper", "research", "学术搜索", "BibTeX", "DOI", or asks to find/search/download academic papers from any source.
---

# Paper Research — 学术文献搜索工具

通过命令行并行搜索多个学术数据库，自动去重合并，下载 PDF，导出 BibTeX。

## Prerequisites

1. CLI binary available (install to `~/.local/bin/` or any directory in `$PATH`):
   ```bash
   scholar-cli --help
   ```

2. For WebBridge sources (Google Scholar, CNKI, WoS), kimi-webbridge daemon must be running:
   ```bash
   ~/.kimi-webbridge/bin/kimi-webbridge status
   ```
   If not running: invoke the `kimi-webbridge` skill.

## 命令速查

```bash
# 英文多源并行搜索（默认 7 个源）
scholar-cli search-en --query "关键词" --limit 10

# 指定源搜索
scholar-cli search-en --query "关键词" --sources openalex,semantic,arxiv,dblp

# Google Scholar（需要 WebBridge + 浏览器）
scholar-cli search-google --query "关键词" --limit 10

# 知网搜索（需要 WebBridge + 知网登录）
scholar-cli search-cnki --query "关键词" --limit 20

# Web of Science（需要 WebBridge + 机构登录）
scholar-cli search-wos --query "关键词" --limit 10

# 按 DOI 查论文详情
scholar-cli detail --doi "10.1145/3442188.3445922"

# 搜索并保存到工作区（自动去重）
scholar-cli search-en --query "关键词" --workspace /path/to/workspace

# 导出 BibTeX
scholar-cli export --workspace /path/to/workspace --output refs.bib

# 检查 WoS 登录状态
scholar-cli login-status --platform wos
```

## 所有可用源

| 源 | 源代码名 | 类型 | 需要登录 |
|---|---------|------|---------|
| OpenAlex | `openalex` | HTTP API | 否 |
| Semantic Scholar | `semantic` | HTTP API | 否 |
| CrossRef | `crossref` | HTTP API | 否 |
| arXiv | `arxiv` | HTTP API | 否 |
| PubMed | `pubmed` | HTTP API | 否 |
| DBLP | `dblp` | HTTP API | 否 |
| bioRxiv/medRxiv | `biorxiv` | HTTP API | 否 |
| Google Scholar | — | WebBridge | 否（但有 CAPTCHA） |
| CNKI 知网 | — | WebBridge | 是 |
| Web of Science | — | WebBridge | 是（机构） |

`search-en` 默认并行搜索前 7 个 HTTP API 源。Google Scholar、CNKI、WoS 需要单独命令调用。

## 领域 → 源映射

不同学科的核心文献分布差异很大，选对源直接决定搜索质量。

| 领域 | 首选源 | 次选源 | 说明 |
|------|--------|--------|------|
| **计算机科学 / AI / ML** | `arxiv`, `semantic`, `dblp` | `openalex` | arXiv 是 CS 预印本主阵地，S2 对 CS 覆盖最好，DBLP 是 CS 文献目录权威 |
| **生物医学 / 临床医学** | `pubmed`, `biorxiv` | `crossref`, `openalex` | PubMed 是 MEDLINE 入口必查，bioRxiv/medRxiv 覆盖生物医学预印本 |
| **物理 / 数学 / 天文** | `arxiv` | `openalex`, `crossref` | 物理学和数学重度依赖 arXiv |
| **社会科学 / 经济学** | `openalex`, `crossref` | `semantic` | 传统期刊为主，OpenAlex 覆盖最广 |
| **工程 / 材料 / 化学** | `crossref`, `openalex` | `semantic` | 传统期刊为主，CrossRef 的 DOI 元数据最全 |
| **中文文献（任何学科）** | `cnki`（知网） | — | 中文期刊、学位论文、会议论文的唯一可靠源。需要登录 |
| **高影响力/引用分析** | `search-wos` | `openalex`, `semantic` | WoS 的引用数据最权威，适合做引文分析和找高影响力论文 |
| **跨学科 / 综述调研** | 全源 `search-en` | `search-google` | 不确定领域时走全源并行，去重后看分布再聚焦 |

### 使用策略

1. **已知领域** → 用 `--sources` 指定 2-3 个核心源，减少噪音
   ```bash
   # AI / CS 论文
   scholar-cli search-en --query "large language model reasoning" --sources arxiv,semantic,dblp --limit 15
   
   # 生物医学论文
   scholar-cli search-en --query "CRISPR gene therapy" --sources pubmed,biorxiv,crossref --limit 15
   
   # 工程/材料
   scholar-cli search-en --query "perovskite solar cell" --sources crossref,openalex --limit 15
   ```

2. **不确定领域** → 先用全源搜一轮，看 `sources` 汇总判断哪些源有料
   ```bash
   scholar-cli search-en --query "quantum computing" --limit 5
   # 看返回的 sources 字段，哪个 count 高就追加搜哪个
   ```

3. **中文研究** → 用 `search-cnki`，确保浏览器已登录知网
   ```bash
   scholar-cli search-cnki --query "大语言模型教育应用" --limit 20
   ```

4. **引用分析/高影响力筛选** → 用 WoS，需要机构 VPN/登录
   ```bash
   scholar-cli search-wos --query "transformer attention" --limit 20
   ```

5. **需要 Google Scholar** → 被 CAPTCHA 拦截时，让用户在浏览器里手动过验证码，之后短时间内可用
   ```bash
   scholar-cli search-google --query "attention mechanism survey"
   ```

## 搜索关键词技巧

- **英文源**：用英文关键词，支持布尔逻辑（取决于各源 API）
- **知网**：用中文关键词，和知网网页搜索行为一致
- **DBLP**：对 CS 会议/期刊论文标题匹配特别准，适合精确查找
- **bioRxiv**：通过 CrossRef DOI 前缀过滤实现，搜索质量与 CrossRef 一致
- **DOI 查详情**：如果从搜索结果里拿到了 DOI，用 `detail` 命令补全完整信息（作者机构、摘要、引用数、PDF 链接）

## 工作区管理

工作区用于跨多次搜索累积论文，自动按 DOI / 标题+作者 去重：

```bash
# 多次搜索存同一个工作区
scholar-cli search-en --query "term1" --workspace ~/research/my-survey --limit 10
scholar-cli search-en --query "term2" --workspace ~/research/my-survey --limit 10

# 最后统一导出 BibTeX
scholar-cli export --workspace ~/research/my-survey --output ~/research/my-survey/refs.bib
```

返回值中 `papers_added` 表示新增了多少篇（已有的不重复添加），`total_stored` 是工作区总数。

## 输出格式

所有命令输出 JSON：
- 成功：`{"ok": true, "data": {...}}`
- 失败：`{"ok": false, "error": {"code": "...", "message": "..."}}`

论文字段：title, authors, abstract, year, doi, venue, volume, issue, pages, citations, references, open_access, pdf_url, source, sources, urls, identifiers

## 常见问题

| 问题 | 原因与解决 |
|------|-----------|
| PubMed 返回 "blocked by NCBI" | NCBI 限制了请求频率或 IP，等一会重试或加 API key |
| DBLP 连接超时 | 部分网络环境下 dblp.org 访问不稳定，不影响其他源 |
| Google Scholar CAPTCHA | 在浏览器里手动过一次验证码即可 |
| 知网 "not_logged_in" | 在浏览器里打开 cnki.net 登录（见下方知网使用说明） |
| 知网滑块验证码 | 不只登录时出现，新会话、频繁请求、自动化检测都会触发（见下方知网使用说明） |
| WoS "not_logged_in" | 需要通过机构 VPN 或 SSO 登录 Web of Science |
| Semantic Scholar 返回 0 结果 | S2 的搜索对关键词敏感，换个更具体的 query 试试 |
| bioRxiv 年份为 0 | 极少数早期论文缺少日期元数据 |
| 某个源超时 | 不影响其他源，结果中 sources 会显示 error 信息 |

## 知网（CNKI）使用说明

使用 `search-cnki` 前，必须在浏览器中完成以下准备：

1. **登录知网** — 在 Chrome 中打开 https://www.cnki.net 并登录账号
2. **过一次滑块验证码** — 打开 https://kns.cnki.net 搜索任意关键词，如果弹出滑块验证码（"安全验证"页面），手动拖动滑块完成验证

知网的滑块验证码（blockPuzzle 类型）不仅在登录时出现，以下场景也会触发：
- 新会话 / 新 IP 首次访问
- 短时间内搜索次数过多
- 检测到自动化浏览器行为

**验证码无法自动破解**，必须手动完成。过一次之后同一浏览器会话可保持一段时间正常使用。

如果 `search-cnki` 返回错误或空结果，优先让用户在浏览器里访问 kns.cnki.net 确认是否需要重新过验证码。

## 依赖

- **7 个 HTTP API 源**（OpenAlex, Semantic Scholar, CrossRef, arXiv, PubMed, DBLP, bioRxiv）：**无需登录**，直接 HTTP 调用
- **Google Scholar**：需要 kimi-webbridge 运行 + Chrome 打开
- **CNKI 知网**：需要 kimi-webbridge 运行 + Chrome 打开 + 知网已登录 + **手动过一次滑块验证码**
- **Web of Science**：需要 kimi-webbridge 运行 + Chrome 打开 + 机构 SSO 登录
