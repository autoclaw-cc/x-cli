# Recipes

English | [中文](./README.md)

Copy-paste-ready scripts. Each one mirrors a scenario in the [main README](../README_EN.md#5-ready-made-scenarios), giving the exact prompt to send to Claude, setup steps, and what to expect back.

## Index

| Recipe | Scenario | Uses |
|---|---|---|
| [plan-kyoto-trip](./plan-kyoto-trip_EN.md) | Plan a full trip | ctrip-cli + booking-cli + travel-planning |
| [find-shanghai-rental](./find-shanghai-rental_EN.md) | Cross-platform rental search (with overseas variants) | 58-cli + anjuke-cli + apartments-cli + rightmove-cli + idealista-cli + xiaohongshu-cli + rental-assistant |
| [gaokao-jiangsu-211](./gaokao-jiangsu-211_EN.md) | Gaokao reach/target/safety list | gaokao-cli + gaokao-assistant |
| [batch-image-shiba-inu](./batch-image-shiba-inu_EN.md) | Batch AI image generation | chatgpt-image-cli or nanobanana-cli |
| [research-local-ai-models](./research-local-ai-models_EN.md) | Research a topic + fetch articles + summarize | google-cli or baidu-cli |

## Contributing a new recipe

Write a new markdown file mirroring the structure of an existing recipe and open a PR. Each should include:

- **When to use**: what real need would lead someone here
- **Setup**: which CLIs to install, which sites to log into
- **Send to Claude**: the copy-paste prompt (code-fenced)
- **What you'll get**: describe the output shape
- **Uses**: CLI and skill links
- **Adapt**: how to tweak for your own scenario
