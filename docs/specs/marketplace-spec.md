# Marketplace Specification

## Host Listing

A host listing represents a machine available for rent.

### Listing Fields
- `display_name`: Human-readable name set by host
- `gpu_model`: GPU name (auto-detected by agent)
- `vram_gb`: GPU VRAM in GB
- `cpu_model`: CPU name
- `ram_gb`: System RAM
- `os`: Operating system
- `price_per_hour_usd`: Host-set hourly rate (min: $0.10, max: $10.00)
- `tags`: Array of tags (e.g., `["gaming", "4k", "rtx4090"]`)
- `online`: Whether agent is connected and ready
- `rating`: Average session rating (0-5 stars)
- `total_sessions`: Lifetime session count

### Browse / Search

Users can filter by:
- GPU VRAM minimum
- Max hourly price
- Tags
- Online now only

Default sort: relevance (online first, then rating, then price)

## Session Booking Flow

1. User opens host detail page
2. User selects duration (15min, 30min, 1hr, 2hr, 4hr, custom)
3. Frontend calls `POST /sessions` → gets `checkout_url`
4. User redirected to Stripe Checkout
5. Payment complete → Stripe webhook → session AUTHORIZED
6. User redirected to `/session/:id` page (session viewer)
7. Viewer shows "Connecting..." while WebRTC handshake completes
8. Stream appears; input is captured

## Session Viewer Requirements

- Full-screen WebRTC video element (16:9, stretches to fit viewport)
- Pointer lock for mouse capture (click to lock, Escape to release)
- Keyboard events forwarded (browser `keydown`/`keyup` → data channel)
- Gamepad API polling at 60Hz for connected controllers
- Live billing timer showing: elapsed time + remaining balance in minutes + USD
- Warning banner at 5 minutes remaining
- "End Session" button (with confirmation dialog)

## Host Dashboard

Hosts can see:
- Current session status (idle / active)
- Session history (date, duration, earnings per session)
- Total earnings (month, all-time)
- Payout history (Stripe Connect transfers)
- Machine uptime and online/offline toggle
- Listing editor (price, display name, tags)

## User Dashboard

Users can see:
- Active session (if any) with link to viewer
- Session history (host, duration, cost per session)
- Wallet balance (current USD balance)
- Top-up wallet button (Stripe Checkout for balance reload)
- Receipt download per session
