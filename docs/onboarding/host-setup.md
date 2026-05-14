# Host Setup Guide

## Requirements

- Windows 10 Pro / Enterprise or Windows 11 (Windows Sandbox requires Pro+)
- NVIDIA, AMD, or Intel GPU (DirectX 12 capable)
- 8 GB RAM minimum (16 GB recommended)
- Virtualization enabled in BIOS (required for Windows Sandbox)
- Stable internet connection (10+ Mbps upload recommended)

## Step 1: Enable Windows Sandbox

Open PowerShell as Administrator:
```powershell
Enable-WindowsOptionalFeature -Online -FeatureName "Containers-DisposableClientVM" -All
```
Restart when prompted.

## Step 2: Install IndraNet Host Agent

Download the installer from the IndraNet dashboard or run the Tauri desktop app.

## Step 3: Register Your Machine

1. Open the IndraNet desktop app
2. Click "Become a Host"
3. Sign in or create an account
4. Complete Stripe Express onboarding (for payouts)
5. Set your price per hour
6. Click "Go Live"

## Step 4: Your Machine is Listed

Once online, your machine appears in the marketplace. You earn 80% of each session's revenue. Payouts via Stripe weekly.

## Troubleshooting

**Windows Sandbox won't start:** Ensure virtualization is enabled in BIOS under CPU settings (Intel VT-x / AMD-V).

**GPU not detected inside sandbox:** Update GPU drivers to latest version. Sandbox requires WDDM 2.5+ drivers.

**Agent won't connect to backend:** Check firewall settings — the agent needs outbound access on port 443 (HTTPS/WSS).
