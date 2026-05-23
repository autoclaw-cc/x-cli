# Find a 2-bedroom in Shanghai's Zhangjiang

English | [中文](./find-shanghai-rental.md)

## When to use

Job change means flat hunting under a budget and a commute constraint. Hit several rental sites in one shot instead of switching between platforms with different filter UIs. Pull tips from Xiaohongshu in the same pass.

## Setup

- [kimi-webbridge](https://www.kimi.com/features/webbridge)
- Download 58-cli and anjuke-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
- Install Xiaohongshu CLI (tips and traps): `brew install xpzouying/agent-cli/xiaohongshu-cli`
- Install the skill: `npx skills add better-world-ai/x-cli --skill rental-assistant`

## Send to Claude

```
Find me a 2-bedroom in Shanghai's Zhangjiang area, monthly rent under ¥5000.
Aim for ≤30 min commute to Zhangjiang Hi-Tech subway, prefer flats with a lift and ensuite.
Hit both 58.com and Anjuke and cross-reference, then check Xiaohongshu for area-specific warnings.
```

## What you'll get

- Cross-platform deduplicated listings (title, price, area, layout, distance, source link)
- Price distribution by complex or sub-neighborhood
- Xiaohongshu posts about the area (warnings, sublease notes, agent reviews)
- Visible anomalies (same listing priced very differently on two sites, obvious phishing posts)

## Uses

[58-cli](../58-cli/) + [anjuke-cli](../anjuke-cli/) + `xiaohongshu-cli` + [rental-assistant](../skills/rental-assistant/) skill

## Adapt

- **Different city**: swap Zhangjiang for Chengdu Chunxi Road, Shenzhen Nanshan, etc.
- **Different layout**: shared room, full 3-bedroom, loft — just say it
- **Overseas**: "1-bedroom in central London, monthly rent under £2000" auto-switches to Rightmove; "1-bedroom in central Madrid" goes to Idealista; "1-bedroom in Cambridge MA" goes to Apartments.com
- **Adding US/UK/EU**: install apartments-cli, rightmove-cli, or idealista-cli first, then name the country/city in your prompt
