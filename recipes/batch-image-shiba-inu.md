# 批量生成柴犬主题图

## 场景

做 PPT 或公众号缺图，想要一组风格一致但场景不同的图。手动开 ChatGPT 或 Gemini 一张一张右键保存太烦，让 AI 批量跑完落到本地。

## 前置

- [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)
- 在 Chrome 里登录 ChatGPT 或 Google Gemini
- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 chatgpt-image-cli 或 nanobanana-cli

## 发给 Claude

```
帮我用 ChatGPT 画 10 张柴犬主题的图，每张换一个场景：
Times Square 穿西装、东京涩谷雨夜、巴黎咖啡馆、京都樱花树下、
冰岛冰川、撒哈拉沙漠、阿尔卑斯滑雪、米其林餐厅、
图书馆读书、海边日落跑步。
所有图竖版 9:16，风格统一为水彩插画。
按 "shiba-{编号}-{场景}.png" 命名落到 ~/Pictures/shiba/ 下。
```

## 你会拿到

- 10 张按命名规则保存到指定目录的 PNG
- 终端里每张图的生成进度（耗时、对应 prompt）
- 如果某张失败（提示词被拒、生成超时），AI 会跳过并最后告诉你哪些没成

## 用到

[chatgpt-image-cli](../chatgpt-image-cli/) 或 [nanobanana-cli](../nanobanana-cli/)

## 改一改

任何主题（产品图、人物、风景、抽象概念）、任何数量（5 张到几十张都行）、任何尺寸都可以。如果想用 Gemini 出图，把 chatgpt-image-cli 换成 nanobanana-cli 即可，提示词不用变。
