# Research the "2025 local AI models" topic

English | [中文](./research-local-ai-models.md)

## When to use

You need to get up to speed on an unfamiliar topic fast — say, what the local AI model landscape looks like in 2025. Reading article after article by hand is slow. Let the AI run the search, pull the bodies, and consolidate.

## Setup

- [kimi-webbridge](https://www.kimi.com/features/webbridge)
- Download google-cli or baidu-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)

## Send to Claude

```
Search "best local AI models 2025" and fetch the body text of the top 10 Google results.
Then do two things:
1. Summarize the consensus points and disagreements in about 200 words
2. List every model mentioned across these articles in a table, sorted by recommendation count

Output in English, with the original links at the end.
```

## What you'll get

- A tight consolidated summary (consensus + disagreements)
- A model comparison table (name, number of articles recommending it, key strengths, typical use cases)
- The original result links so you can drill in
- If any pages fail to scrape (paywall, 403), the AI flags which ones were skipped

## Uses

[google-cli](../google-cli/) or [baidu-cli](../baidu-cli/)

## Adapt

- **Chinese topics**: swap google-cli for baidu-cli
- **Result count**: top 5 or top 20, your call
- **Constraints**: last 30 days only, exclude certain domains, sort by date
- **Output format**: summary, comparison table, SWOT, Q&A FAQ, or slide outline — just ask
