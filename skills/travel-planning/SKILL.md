---
name: travel-planning
description: Use when the user wants to plan a trip, compare hotels/flights/attractions across platforms, or asks for travel recommendations. Invoke when user mentions "旅游攻略", "旅行规划", "travel planning", or asks to plan a trip to any destination.
---

# Travel Planning

Orchestrates 3 CLI tools to help users plan domestic and international trips — covering hotels, flights, attractions, destination guides, and user reviews.

## Prerequisites

1. kimi-webbridge daemon running:
   ```bash
   ~/.kimi-webbridge/bin/kimi-webbridge status
   ```
   If not running: invoke the `kimi-webbridge` skill.

2. CLI binaries available (install to `~/.local/bin/` or anywhere on `$PATH`):
   ```bash
   ctrip-cli --help
   booking-cli --help
   xiaohongshu-cli --help
   ```

## CLI Capabilities

| Capability | ctrip-cli | booking-cli | xiaohongshu-cli |
|-----------|-----------|-------------|-----------------|
| Hotel search | Domestic + International | International only | - |
| Flight search | Domestic + International | - | - |
| Attractions | Yes (with scores, prices) | - | - |
| Destination guide | Must-do lists, travel notes | - | - |
| User reviews/tips | - | - | Travel notes search |
| Login required | No (member prices need login) | No | Read: No, Write: Yes |
| CAPTCHA risk | Low | Low | Low |

## Workflow: Domestic Trip (国内游)

Use ctrip + xiaohongshu. Example: planning a trip to Chengdu.

```bash
# 1. Get destination overview from Ctrip (must-do sights, restaurants, shops, travel notes)
ctrip-cli destination --name chengdu104

# 2. Search user travel tips on Xiaohongshu
xiaohongshu-cli search "成都旅游攻略" --limit 10

# 3. Search hotels on Ctrip
ctrip-cli search-hotels --keyword "成都" --checkin 2026/07/01 --checkout 2026/07/05 --limit 10

# 4. Search flights
ctrip-cli search-flights --from SHA --to CTU --date 2026-07-01 --limit 10

# 5. Browse attractions
ctrip-cli search-attractions --destination chengdu104 --limit 10
```

## Workflow: International Trip (出境游)

Use all 3 CLIs. Example: planning a trip to Kyoto.

```bash
# 1. Get destination overview from Ctrip
ctrip-cli destination --name kyoto430

# 2. Search user travel tips on Xiaohongshu
xiaohongshu-cli search "京都旅游攻略" --limit 10

# 3. Compare hotels: Ctrip vs Booking.com
ctrip-cli search-hotels --keyword "京都" --city-id 734 --country-id 78 --checkin 2026/07/01 --checkout 2026/07/05 --limit 10
booking-cli search-hotels --destination "Kyoto" --checkin 2026-07-01 --checkout 2026-07-05 --limit 10

# 4. Search flights
ctrip-cli search-flights --from SHA --to KIX --date 2026-07-01 --limit 10

# 5. Browse attractions
ctrip-cli search-attractions --destination kyoto430 --limit 10
```

## Quick Command Reference

### ctrip-cli

```bash
ctrip-cli destination --name SLUG
ctrip-cli search-hotels --keyword "城市" [--city-id N --country-id N] [--checkin YYYY/MM/DD] [--checkout YYYY/MM/DD] [--limit N]
ctrip-cli search-flights --from IATA --to IATA [--date YYYY-MM-DD] [--limit N]
ctrip-cli search-attractions --destination SLUG [--limit N]
```

### booking-cli

```bash
booking-cli search-hotels --destination "City" [--checkin YYYY-MM-DD] [--checkout YYYY-MM-DD] [--adults N] [--rooms N] [--limit N]
```

### xiaohongshu-cli

```bash
xiaohongshu-cli search "关键词" [--limit N]
xiaohongshu-cli view <note_id> <xsec_token>
xiaohongshu-cli user <user_id>
xiaohongshu-cli feeds [--limit N]
xiaohongshu-cli login-status
xiaohongshu-cli screenshot <note_id> <xsec_token> [-o path.png]
```

## Destination ID Cheat Sheet

| Destination | ctrip slug | booking name |
|------------|-----------|-------------|
| 上海 | shanghai2 | Shanghai |
| 北京 | beijing1 | Beijing |
| 杭州 | hangzhou14 | Hangzhou |
| 成都 | chengdu104 | Chengdu |
| 西安 | xian7 | Xi'an |
| 三亚 | sanya61 | Sanya |
| 厦门 | xiamen21 | Xiamen |
| 广州 | guangzhou152 | Guangzhou |
| 重庆 | chongqing158 | Chongqing |
| 东京 | tokyo294 | Tokyo |
| 京都 | kyoto430 | Kyoto |
| 大阪 | osaka293 | Osaka |
| 首尔 | seoul234 | Seoul |
| 曼谷 | bangkok191 | Bangkok |
| 新加坡 | singapore53 | Singapore |
| 巴黎 | paris308 | Paris |
| 伦敦 | london309 | London |

### Flight City Codes (IATA)

上海: SHA, 北京: BJS, 广州: CAN, 深圳: SZX, 成都: CTU, 杭州: HGH, 西安: SIA, 三亚: SYX, 重庆: CKG, 南京: NKG, 东京: TYO, 大阪: KIX, 首尔: SEL, 曼谷: BKK, 新加坡: SIN

## Login & CAPTCHA Notes

| Platform | Login needed | Without login | CAPTCHA risk |
|----------|-------------|---------------|-------------|
| Ctrip | Optional | All features work; prices may differ (no member discount) | Low — never triggered during testing |
| Booking.com | Optional | Search works; may show "Verify email" popup if previously logged in — complete in Chrome and retry | Low |
| Xiaohongshu | Read: No, Write: Yes | search, view, feeds, user all work. like/comment/post require login | Low |

## Tips

- **Start with destination overview**: Run `ctrip-cli destination` first to understand the destination, then drill into hotels/flights/attractions.
- **International hotel comparison**: Always run both `ctrip-cli search-hotels` and `booking-cli search-hotels` for international destinations — prices and availability differ significantly.
- **Ctrip international hotels**: Use `ctrip-cli destination` first to get the `hotel_url`, extract `city-id` and `country-id` from it, then pass to `search-hotels`. Keyword-only search may return wrong city for international destinations.
- **Xiaohongshu for real tips**: Search Xiaohongshu for first-person travel experiences, restaurant recommendations, and hidden gems that structured platforms miss.
- **Date formats differ**: Ctrip hotels use `YYYY/MM/DD`, Ctrip flights use `YYYY-MM-DD`, Booking uses `YYYY-MM-DD`.
- **Xiaohongshu search retry**: First search call may fail if no browser tab exists yet — simply retry once.
- **Output format**: All 3 CLIs return `{"ok": true, "data": ...}` on success and `{"ok": false, "error": {...}}` on failure.
