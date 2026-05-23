# Plan a Kyoto trip in one sentence

English | [中文](./plan-kyoto-trip.md)

## When to use

You suddenly feel like getting out of town but don't want to burn an afternoon comparing hotels, flights, and attractions. Let the AI do that work and hand you back an actionable itinerary.

## Setup

- [kimi-webbridge](https://www.kimi.com/features/webbridge) (drives your local Chrome)
- Download ctrip-cli and booking-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
- Install the skill: `npx skills add better-world-ai/x-cli --skill travel-planning`

## Send to Claude

```
Plan a 5-day trip to Kyoto in June, budget under $2000.
I like kimono experiences, specialty coffee shops, and serious temple visits — not shopping.
For hotels, pick two with ratings 8.5+ near a subway station, comparing Ctrip (domestic) vs. Booking (international) and taking the lower price.
Flights from Shanghai, no red-eyes.
```

## What you'll get

- Recommended flights (airline, time slots, price range, layovers)
- 2-3 candidate hotels (rating, location, Ctrip vs. Booking price)
- 5-day daily itinerary (sights, route, estimated duration, opening hours)
- Total budget breakdown (flights + hotel + entry fees + food estimate)

The AI pauses at key checkpoints to confirm adjustments rather than blasting through end-to-end.

## Uses

[ctrip-cli](../ctrip-cli/) + [booking-cli](../booking-cli/) + [travel-planning](../skills/travel-planning/) skill

## Adapt

Swap Kyoto for any city, change the dates and budget to your real numbers, switch the preferences to yours. Traveling with kids? Add child-friendly attraction weighting. Honeymoon? Add couple-oriented restaurant weighting. Just say it.
