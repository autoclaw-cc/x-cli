# Jiangsu candidate, 580 points — which 211 schools?

English | [中文](./gaokao-jiangsu-211.md)

## When to use

Scores are out and the application form is due in three days. Your provincial rank sits in some awkward middle bracket, and you need to surface the "reach", "target", and "safety" lists fast.

## Setup

- [kimi-webbridge](https://www.kimi.com/features/webbridge)
- Download gaokao-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
- Install the skill: `npx skills add better-world-ai/x-cli --skill gaokao-assistant`

## Send to Claude

```
I'm a 2026 Jiangsu physics-track candidate, score 580, provincial rank around 35,000.
List the 211 schools I can apply to, sorted into "reach", "target", and "safety".
I prefer computer science, automation, and electronic information majors; regional preference is the Yangtze River Delta and South China.
For each school, give the lowest admission rank over the past three years and the corresponding majors so I can compare.
```

## What you'll get

- Three-tier school list: reach (rank higher than historical cutoff), target (rank close), safety (rank clearly below)
- Each school's last-three-year admission ranks, major groups, and corresponding majors
- How your preferred majors fit at each school (admission situation and rank threshold)
- Side-by-side comparison within each tier, sorted by your region and major preference

The AI doesn't make the final choice, but every piece of data you need to decide ends up on one screen.

## Uses

[gaokao-cli](../gaokao-cli/) + [gaokao-assistant](../skills/gaokao-assistant/) skill

## Adapt

Swap province, score, preferred majors, or tier (Double First-Class / 985 / non-211) freely. If you're unsure of your rank, ask first ("what rank is this score roughly?"). For major-group breakdowns, follow up ("compare the computer science major groups across these schools").
