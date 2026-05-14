# PoC 04: Stripe Payment Gate

## Goal
A Stripe test payment gates access to a mock WebRTC stream. User pays $1 → receives a JWT session token → can connect to a stream endpoint.

## Success Criteria
- [ ] POST /create-checkout creates a Stripe Checkout session
- [ ] User completes payment on Stripe's hosted checkout page
- [ ] Stripe sends `checkout.session.completed` webhook
- [ ] Server issues a JWT with `session_id` claim
- [ ] Client receives JWT and displays the stream connection info
- [ ] Invalid/expired JWT rejected with 401

## Prerequisites
- Stripe account with test mode enabled
- Stripe CLI for webhook forwarding: `stripe listen --forward-to localhost:8080/webhook`

## How to Run

1. Copy `.env.example` to `.env` and add your Stripe test keys
2. Start the server:
   ```bash
   cd server
   go run main.go
   ```
3. Open `client/index.html` in a browser
4. Click "Pay $1 to Connect"
5. Use Stripe test card: `4242 4242 4242 4242`, any future date, any CVC
6. Verify JWT is shown and "Connect to Stream" button appears

## Test Cards
- Success: `4242 4242 4242 4242`
- Decline: `4000 0000 0000 0002`
- 3D Secure: `4000 0000 0000 3220`

## Results

TODO: Fill in after running.

| Test | Result |
|------|--------|
| Checkout session created | TODO |
| Payment completed (test card) | TODO |
| Webhook received + verified | TODO |
| JWT issued | TODO |
| JWT verified on connect | TODO |
| Expired JWT rejected | TODO |
