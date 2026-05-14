# Web — Next.js 14 Marketplace Frontend

## Overview
The IndraNet web marketplace built with Next.js 14 App Router, TypeScript, and Tailwind CSS.

## Pages
- `/` — Browse available hosts
- `/hosts/[id]` — Host detail + book session
- `/session/[id]` — Active session viewer (WebRTC stream)
- `/dashboard` — User dashboard (sessions, wallet)
- `/dashboard/host` — Host dashboard (earnings, listings)

## Running
```bash
pnpm dev
```

## Environment Variables
See `../../.env.example` — set `NEXT_PUBLIC_*` vars.
