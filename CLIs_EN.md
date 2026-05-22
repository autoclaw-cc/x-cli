# Included CLIs

English | [中文](./CLIs.md)

Grouped by scenario. Each CLI builds and runs independently — see the README inside each directory for usage details.

## 🔍 Search & scraping

- [`baidu-cli`](./baidu-cli/) — Baidu search.
- [`google-cli`](./google-cli/) — Google search + fetch the target page's content.

## 🎨 Image generation

- [`nanobanana-cli`](./nanobanana-cli/) — Generate images with Google Gemini 2.5 Flash Image (Nano Banana).
- [`chatgpt-image-cli`](./chatgpt-image-cli/) — Generate images via the ChatGPT web app.

## ✈️ Travel planning

- [`ctrip-cli`](./ctrip-cli/) — Search hotels, flights, attractions, and destination guides on Ctrip.
- [`booking-cli`](./booking-cli/) — Search international hotels on Booking.com.
- Companion skill: [`travel-planning`](./skills/travel-planning/) — lets the agent compose the two tools above into an end-to-end trip plan.

## 🏠 Rental listings

- [`58-cli`](./58-cli/) — 58.com (mainland China).
- [`anjuke-cli`](./anjuke-cli/) — Anjuke (mainland China).
- [`apartments-cli`](./apartments-cli/) — Apartments.com (US).
- [`rightmove-cli`](./rightmove-cli/) — Rightmove (UK).
- [`idealista-cli`](./idealista-cli/) — Idealista (Spain / Italy / Portugal).
- Companion skill: [`rental-assistant`](./skills/rental-assistant/) — compare rental listings across platforms.

## 🎓 Gaokao (college admissions in China)

- [`gaokao-cli`](./gaokao-cli/) — Look up score lines, colleges, and majors.
- Companion skill: [`gaokao-assistant`](./skills/gaokao-assistant/) — gaokao admissions decision assistant.
