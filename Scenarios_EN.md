# What you can do

English | [中文](./Scenarios.md)

Five scenarios already wired up — say what you want in plain language and the AI handles the rest. The CLIs are the tools underneath; the AI composes them for you.

## ✈️ Plan a full trip in one sentence

> "Plan a 5-day trip to Kyoto in June"

The AI runs Ctrip and Booking.com side-by-side — compares domestic vs. international hotel rates, books flights, lists must-see attractions, and hands you a ready-to-execute itinerary.

Uses: [ctrip-cli](./ctrip-cli/) · [booking-cli](./booking-cli/) · [travel-planning](./skills/travel-planning/)

## 🏠 Search multiple rental sites at once

> "Find me a 1-bedroom in central London under £2000/month"

The AI queries 58.com, Anjuke, Apartments.com, Rightmove, and Idealista in parallel — filters by price, layout, and location, then gives you a cross-platform comparison. Works whether you're hunting domestic shares or long-term overseas rentals.

Uses: [58-cli](./58-cli/) · [anjuke-cli](./anjuke-cli/) · [apartments-cli](./apartments-cli/) · [rightmove-cli](./rightmove-cli/) · [idealista-cli](./idealista-cli/) · [rental-assistant](./skills/rental-assistant/)

## 🎓 Gaokao admissions assistant

> "I'm a Jiangsu test-taker with 580 points — which 211 schools can I get into?"

The AI pulls official score lines and prior-year admissions data, factors in your provincial rank and preferences, and gives you a tiered list of reach / target / safety schools.

Uses: [gaokao-cli](./gaokao-cli/) · [gaokao-assistant](./skills/gaokao-assistant/)

## 🎨 Generate images with AI

> "Draw a shiba inu in a suit, standing in Times Square"

The AI generates an image via the ChatGPT web app or Google Gemini and saves it locally — no copy-paste, no API key registration, runs in your already-logged-in Chrome.

Uses: [chatgpt-image-cli](./chatgpt-image-cli/) · [nanobanana-cli](./nanobanana-cli/)

## 🔍 Search + scrape pages

> "Search Google for X, then fetch the body text of the top 10 results"

The AI runs Google or Baidu search, follows each result, and pulls the full content back. A building block for more complex pipelines.

Uses: [google-cli](./google-cli/) · [baidu-cli](./baidu-cli/)

---

Want a new scenario? Use the [`agent-cli-creator`](https://github.com/better-world-ai/agent-cli-creator) skill — one plain-language sentence and the AI builds you a new CLI. See "Build your own CLI" in the main README.
