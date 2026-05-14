# Shadow PC Analysis

## What They Do
Cloud gaming PC subscription service. You get a full Windows 10 VM in a data center with a gaming GPU, accessible via their streaming app.

## Strengths
- Full Windows desktop (not just games — video editing, any application)
- Good streaming quality
- Subscription model provides revenue predictability

## Weaknesses
- Expensive ($29.99/mo minimum)
- Centralized data centers (higher latency for some regions)
- Fixed resource tiers (can't choose GPU)
- No marketplace — you can't rent someone's home GPU

## Technical Note
Shadow uses proprietary streaming (not WebRTC). Latency is acceptable (~30-50ms in good conditions) but not as low as Parsec.

## Relevance
Shadow proves there's a consumer market for "cloud gaming PC" beyond just streaming game titles. Our advantage: peer hardware = lower prices + geographic diversity + any GPU.
