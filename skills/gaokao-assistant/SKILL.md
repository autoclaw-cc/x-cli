---
name: gaokao-assistant
description: Use when the user asks about 高考 (Chinese college entrance exam) data — score lines, one-score-one-rank tables, university info, or admission guidance. Invoke when user mentions "高考", "分数线", "一分一段", "志愿填报", "大学排名", "985", "211", or asks about college admission in China.
---

# gaokao-cli

Query Chinese college entrance exam (高考) data from the command line. Data source: 掌上高考 (gaokao.cn), operated by China Education Online (中国教育在线), under the Ministry of Education.

## Prerequisites

`gaokao-cli` must be installed at `~/.local/bin/gaokao-cli`. No WebBridge or browser required — all data comes from public CDN JSON APIs.

## Commands

### 1. List provinces

```bash
gaokao-cli provinces
```

Returns all 31 provinces with IDs. Use province names (e.g. `北京`) or IDs (e.g. `11`) in other commands.

### 2. Score lines (省控线/批次线)

```bash
gaokao-cli score-line --province 北京
gaokao-cli score-line --province 北京 --year 2025
gaokao-cli score-line --province 河南 --year 2025 --type 物理类
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--province` | `-p` | Yes | Province name or ID |
| `--year` | `-y` | No | Year (default: latest) |
| `--type` | `-t` | No | 综合/理科/文科/物理类/历史类 |

### 3. Score section (一分一段表)

```bash
# Full table
gaokao-cli score-section --province 北京 --year 2025

# Lookup specific score — returns rank, count, and 同位分 (equivalent scores in prior years)
gaokao-cli score-section --province 天津 --year 2025 --score 650

# Old gaokao provinces with 文理分科
gaokao-cli score-section --province 河南 --year 2024 --type 理科 --score 600
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--province` | `-p` | Yes | Province name or ID |
| `--year` | `-y` | No | Year (default: latest) |
| `--type` | `-t` | No | 综合/理科/文科/物理类/历史类 |
| `--level` | `-l` | No | 本科/专科 |
| `--score` | `-s` | No | Specific score to lookup |

**Key feature:** Each entry includes `appositive_fraction` — equivalent scores in previous years for the same rank (同位分), invaluable for cross-year comparison.

### 4. School search (院校查询)

```bash
gaokao-cli school --name 清华
gaokao-cli school --985
gaokao-cli school --211 --province 北京
gaokao-cli school --dual-class --province 上海
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--name` | `-n` | No | Fuzzy name search |
| `--province` | `-p` | No | Filter by province |
| `--985` | | No | 985 universities only |
| `--211` | | No | 211 universities only |
| `--dual-class` | | No | 双一流 universities only |
| `--level` | `-l` | No | 本科 or 专科 |

### 5. Batch reform history (批次改革历史)

```bash
gaokao-cli batch-history --province 北京
gaokao-cli batch-history   # all provinces
```

Shows when provinces merged batches or switched from 文理分科 to new gaokao model. Essential context for interpreting historical score lines.

## Output format

All commands output JSON to stdout:
- Success: `{"ok": true, "data": ...}`
- Error: `{"ok": false, "error": {"code": "...", "message": "..."}}`

## Data coverage

- **31 provinces** across mainland China
- **Score lines:** 2014–2025 (varies by province)
- **一分一段表:** 2016–2025 (varies by province)
- **Schools:** ~2,957 institutions including 985/211/双一流 tags

## Province subject types

Different provinces use different subject classification:

| Type | Description | Used by |
|------|-------------|---------|
| 综合 | Comprehensive (3+3 model) | 北京, 天津, 上海, 浙江, 山东, 海南 |
| 物理类/历史类 | Physics/History track (3+1+2 model) | 广东, 湖北, 湖南, 河北, 福建, 重庆, 辽宁, 江苏 + 2025新增 |
| 理科/文科 | Traditional Science/Arts | Remaining provinces (shrinking as reform spreads) |

Use `batch-history` to check when a province switched models.
