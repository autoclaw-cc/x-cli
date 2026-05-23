# 江苏考生 580 分能上哪些 211

[English](./gaokao-jiangsu-211_EN.md) | 中文

## 场景

成绩出来，志愿表三天后就要交。位次在中段，往年录取数据散在各处，需要快速把「能冲、可稳、保底」三档拉出来。

## 前置

- [kimi-webbridge](https://www.kimi.com/zh-cn/features/webbridge)
- 从 [Releases](https://github.com/better-world-ai/x-cli/releases) 下载并安装 gaokao-cli
- 装 skill：`npx skills add better-world-ai/x-cli --skill gaokao-assistant`

## 发给 Claude

```
我是 2026 年江苏物理类考生，分数 580，省排名 35000 左右。
帮我列出能报的 211 院校，按"冲、稳、保"三档分。
偏好计算机、自动化、电子信息方向，地域优先长三角和华南。
每所学校给出最近三年的最低录取位次和对应专业，方便对照判断。
```

## 你会拿到

- 三档院校清单：冲（位次高于历年录取线）、稳（位次接近）、保（位次明显低于）
- 每所学校近三年的录取位次、专业组、对应专业
- 你的偏好专业在该校的招生情况和位次要求
- 同档院校的对比表（按你的偏好地域和专业排序）

AI 不替你做最终选择，但你做选择需要的所有数据都在一个清单里。

## 用到

[gaokao-cli](../gaokao-cli/) + [gaokao-assistant](../skills/gaokao-assistant/) skill

## 改一改

换省份、换分数、换偏好专业、换院校档次（双一流、985、双非）都可以直接说。如果对位次没把握，可以让它先帮你换算（「我这个分大概是什么位次」）；如果想看专业组细节，可以再追问（「计算机类专业组在这几所里的差异」）。
