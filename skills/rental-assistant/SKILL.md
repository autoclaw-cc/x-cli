---
name: rental-assistant
description: Use when the user wants to search for rental apartments, compare listings across platforms, or asks for help finding housing. Invoke when user mentions "租房", "找房", "房源", "rent", "apartment search", or asks to find housing in any supported region.
---

# 租房搜索助手

跨平台租房搜索工具，覆盖中国、美国、英国、西欧（西班牙/意大利/葡萄牙）。通过 5 个 CLI 统一查询各大租房平台，返回结构化 JSON。

## 前置条件

kimi-webbridge daemon 必须运行：
```bash
~/.kimi-webbridge/bin/kimi-webbridge status
```
未运行则先调用 `kimi-webbridge` skill 启动。

## CLI 一览

| CLI | 平台 | 覆盖范围 | 需要登录 |
|-----|------|---------|---------|
| `58-cli` | 58同城 | 中国大陆 | 否 |
| `anjuke-cli` | 安居客 | 中国大陆 | 否 |
| `apartments-cli` | Apartments.com | 美国全境 | 否 |
| `rightmove-cli` | Rightmove | 英国全境 | 否 |
| `idealista-cli` | Idealista | 西班牙/意大利/葡萄牙 | 否 |

**辅助工具**: `xiaohongshu-cli` 可搜索租房攻略和避坑经验（非结构化房源）。通过 Homebrew 安装：

```bash
brew install xpzouying/agent-cli/xiaohongshu-cli
```

（一次性 tap：`brew tap xpzouying/agent-cli`。覆盖 macOS arm64/amd64 和 Linux amd64/arm64）

所有 CLI 输出统一格式：`{"ok": true, "data": {...}}` / `{"ok": false, "error": {"code": "...", "message": "..."}}`

## 按地区选择 CLI

| 用户说 | 用什么 |
|--------|--------|
| 中国任意城市租房 | 58-cli + anjuke-cli（两个都查，对比） |
| 美国租房 | apartments-cli |
| 英国租房 | rightmove-cli |
| 西班牙/意大利/葡萄牙租房 | idealista-cli |
| "看看大家怎么说" / 避坑 | xiaohongshu-cli search "XX租房" |

## 命令参考

### 58-cli

```bash
# 搜索
58-cli search --city sz --keyword 南山 --limit 10
58-cli search --city sh --keyword 浦东 --min-price 3000 --max-price 5000 --limit 10

# 详情
58-cli detail --url "https://sz.58.com/zufang/xxxxx.shtml"
```

| 参数 | 说明 |
|------|------|
| `--city` | 城市缩写（sz=深圳, sh=上海, bj=北京, gz=广州, cd=成都, hz=杭州, nj=南京, wh=武汉, cq=重庆, xa=西安, tj=天津, su=苏州） |
| `--keyword` | 区域/商圈/小区名 |
| `--min-price` / `--max-price` | 月租范围 |
| `--limit` | 返回数量（默认 20） |

### anjuke-cli

```bash
# 搜索
anjuke-cli search --city sz --keyword 南山 --limit 10
anjuke-cli search --city sh --keyword 张江 --min-price 3000 --max-price 6000 --limit 10

# 详情
anjuke-cli detail --city sz --id 123456789
```

| 参数 | 说明 |
|------|------|
| `--city` | 城市两字母代码（sz=深圳, bj=北京, sh=上海, gz=广州, hz=杭州, cd=成都, tj=天津, nj=南京, wh=武汉, cs=长沙, cq=重庆, xa=西安） |
| `--keyword` | 区域/商圈/小区名 |
| `--min-price` / `--max-price` | 月租范围 |
| `--limit` | 返回数量（默认 20） |

### apartments-cli（美国）

```bash
# 搜索
apartments-cli search --location new-york-ny --min-beds 2 --max-price 3000 --limit 10
apartments-cli search --location san-francisco-ca --min-beds 1 --max-beds 2 --min-price 1500 --max-price 2500 --limit 10
apartments-cli search --location chicago-il --page 2 --limit 10

# 详情
apartments-cli detail --url "https://www.apartments.com/xxx/yyy/"
```

| 参数 | 说明 |
|------|------|
| `--location` | 城市 slug（格式: city-name-state，如 new-york-ny, los-angeles-ca, seattle-wa） |
| `--min-beds` / `--max-beds` | 卧室数 |
| `--min-price` / `--max-price` | 月租范围（美元） |
| `--limit` | 返回数量（默认 20） |
| `--page` | 翻页 |

**常用 location slug**: new-york-ny, los-angeles-ca, san-francisco-ca, chicago-il, seattle-wa, boston-ma, washington-dc, austin-tx, denver-co, miami-fl, san-diego-ca, portland-or

### rightmove-cli（英国）

```bash
# 搜索
rightmove-cli search --location London --min-beds 2 --max-price 2000 --limit 10
rightmove-cli search --location Manchester --min-beds 1 --max-beds 2 --limit 10

# 详情
rightmove-cli detail --url "https://www.rightmove.co.uk/properties/xxxxx"
```

| 参数 | 说明 |
|------|------|
| `--location` | 城市名（London, Manchester, Birmingham, Edinburgh, Bristol, Leeds, Liverpool, Glasgow, Cambridge, Oxford） |
| `--min-beds` / `--max-beds` | 卧室数 |
| `--min-price` / `--max-price` | 月租范围（英镑） |
| `--limit` | 返回数量（默认 20） |
| `--page` | 翻页 |

### idealista-cli（西欧）

```bash
# 搜索
idealista-cli search --country spain --city madrid-madrid --limit 10
idealista-cli search --country italy --city roma --limit 10
idealista-cli search --country portugal --city lisboa --limit 10

# 详情
idealista-cli detail --country spain --url "https://www.idealista.com/inmueble/xxxxx/"
```

| 参数 | 说明 |
|------|------|
| `--country` | spain / italy / portugal |
| `--city` | 城市 slug（spain: madrid-madrid, barcelona-barcelona, valencia, sevilla; italy: roma, milano, firenze; portugal: lisboa, porto） |
| `--limit` | 返回数量（默认 20） |

> 价格/房间数过滤暂未实现（idealista 用 path-based 过滤，需要按 country 分别编码）。需要过滤先全量拉取后在调用方筛选。

### xiaohongshu-cli（辅助）

```bash
xiaohongshu-cli search "上海租房避坑" --limit 10
xiaohongshu-cli search "纽约租房攻略" --limit 10
xiaohongshu-cli view <note_id> <xsec_token>
```

## 工作流

### 中国城市租房

```bash
# 1. 两个平台同时搜（对比房源和价格）
58-cli search --city sz --keyword 南山 --min-price 3000 --max-price 5000 --limit 10
anjuke-cli search --city sz --keyword 南山 --min-price 3000 --max-price 5000 --limit 10

# 2. 看中的房源查详情
58-cli detail --url "https://sz.58.com/zufang/xxxxx.shtml"
anjuke-cli detail --city sz --id 123456789

# 3. 搜避坑经验
xiaohongshu-cli search "深圳南山租房避坑" --limit 5
```

### 美国租房

```bash
# 1. 搜索房源
apartments-cli search --location new-york-ny --min-beds 1 --max-price 2500 --limit 10

# 2. 查详情（含户型定价表、设施、图片）
apartments-cli detail --url "https://www.apartments.com/xxx/yyy/"

# 3. 搜攻略
xiaohongshu-cli search "纽约租房经验" --limit 5
```

### 英国 / 欧洲租房

```bash
# 英国
rightmove-cli search --location London --min-beds 1 --max-price 1800 --limit 10
rightmove-cli detail --url "https://www.rightmove.co.uk/properties/xxxxx"

# 西班牙
idealista-cli search --country spain --city barcelona-barcelona --limit 10
idealista-cli detail --country spain --url "https://www.idealista.com/inmueble/xxxxx/"
```

## 选房注意事项

搜到房源后，帮用户从以下角度做分析和提醒：

### 通勤

- 问清用户上班/上学地点，估算通勤时间
- 地铁直达 ≤30min 是舒适区，超过 45min 或换乘 2 次以上要提醒用户权衡
- 美国看开车通勤，注意高峰堵车；英国/欧洲看公共交通

### 性价比

- 同区域多搜几套对比，感受当地租金水位
- 中国：58 和安居客的价格可能不同，以实际为准
- 美国：apartments.com 标价通常是 range，实际入住价看详情里的 unit pricing
- 价格明显低于同区域均价的要警惕是否虚假房源

### 房屋状况

- 详情页有图片的优先看，注意装修新旧、采光、家具配置
- 中国关注：朝向（南向优先）、楼层（避开一楼和顶楼）、是否有独立卫生间
- 英国关注：EPC 能效等级、是否含 council tax、bills included
- 美国关注：sqft 面积、是否含水电、停车位

### 安全与环境

- 中国：封闭小区 > 开放式，有物业管理 > 无，独门独户 > 合租混住
- 可以用 `xiaohongshu-cli search "{小区名} 避坑"` 搜真实反馈
- 美国：注意周边 neighborhood safety，可提醒用户查 crime map
- 英国：注意 council 区域，不同区域治安差异大

### 中介与合同

- 中国：房东直租 > 正规中介 > 二房东，信息矛盾的要警惕
- 注意押付方式：押一付一最灵活，押二付三 / 付半年要考虑风险
- 英国：注意 agency fees 和 deposit protection scheme
- 美国：注意 application fee、security deposit、lease break penalty
- 西欧：注意 fianza（押金）通常 1-2 个月

### 合同灵活性

- 起租期长短、能否提前退租、违约金多少
- 短租需求（<6个月）优先找明确支持短租的房源
- 留学生注意合同期限是否匹配签证/学期

### 信息真实性

- 价格过低、图片过度精修、多平台重复发布 → 可能是引流假房源
- 联系方式只留微信不留电话 → 可能是二房东
- 挂牌时间很长还没租出去 → 可能有隐藏问题，值得追问原因

## CAPTCHA 与限频

| 平台 | CAPTCHA 风险 | 处理方式 |
|------|-------------|---------|
| 58同城 | 低 | 正常使用即可 |
| 安居客 | 低 | 正常使用即可 |
| Apartments.com | 低 | 偶尔触发，在 Chrome 完成验证后重试 |
| Rightmove | 低 | 正常使用即可 |
| Idealista | 低-中 | 高频访问可能触发，控制请求间隔 |

## Tips

- **输出必须带链接**: 整理房源清单时，每一条都必须附上原始 URL（搜索结果里的 url 字段），方便用户直接打开查看和联系
- **中国城市同时查两个平台**: 58 和安居客房源不完全重叠，两个都搜能看到更多选项
- **先搜再看详情**: search 拿到列表后，挑感兴趣的再 detail，避免浪费请求
- **xiaohongshu 是好帮手**: 搜 "{城市}租房避坑" 可以获取真实租客反馈，帮助判断
- **注意货币单位**: 中国=人民币，美国=美元，英国=英镑，西欧=欧元
- **租房 CLI 不需要登录**: 5 个租房 CLI 直接用，无需任何账号
- **WebBridge 必须运行**: 所有 CLI 依赖 kimi-webbridge daemon（端口 10086）
