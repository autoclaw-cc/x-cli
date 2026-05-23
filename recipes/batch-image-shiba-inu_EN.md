# Batch shiba inu images for a deck

English | [中文](./batch-image-shiba-inu.md)

## When to use

You need a set of images for a slide deck or post — consistent style, different scenes. Manually opening ChatGPT or Gemini, right-click-saving each one is fine for one, painful for thirty. Let the AI batch them to local disk.

## Setup

- [kimi-webbridge](https://www.kimi.com/features/webbridge)
- Be signed into ChatGPT or Google Gemini in Chrome
- Download chatgpt-image-cli or nanobanana-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)

## Send to Claude

```
Use ChatGPT to generate 10 shiba inu themed images, each in a different scene:
Times Square in a suit, Tokyo Shibuya on a rainy night, Paris café, under Kyoto cherry blossoms,
Iceland glacier, Sahara desert, Alps skiing, Michelin restaurant,
reading in a library, jogging by the beach at sunset.
All vertical 9:16, unified watercolor illustration style.
Save to ~/Pictures/shiba/ with naming "shiba-{number}-{scene}.png".
```

## What you'll get

- 10 PNGs saved to the directory under the given naming pattern
- Per-image progress in the terminal (elapsed time, the prompt used)
- If any image fails (rejected prompt, generation timeout), the AI skips it and lists what didn't make it at the end

## Uses

[chatgpt-image-cli](../chatgpt-image-cli/) or [nanobanana-cli](../nanobanana-cli/)

## Adapt

Any subject (product shots, characters, scenery, abstract concepts), any count (5 to several dozen), any size. To use Gemini instead of ChatGPT, swap chatgpt-image-cli for nanobanana-cli — the prompt stays the same.
