import { NextRequest, NextResponse } from 'next/server'

// Stripe webhooks are handled by the Go backend directly.
// This Next.js route exists only for local development convenience
// where the Go backend may not be running.
//
// In production: Stripe sends webhooks to https://api.indranet.io/v1/webhooks/stripe
export async function POST(req: NextRequest) {
  // Forward to backend
  const backendUrl = process.env.NEXT_PUBLIC_API_URL
  if (!backendUrl) {
    return NextResponse.json({ error: 'backend URL not configured' }, { status: 500 })
  }

  const body = await req.arrayBuffer()
  const response = await fetch(`${backendUrl}/v1/webhooks/stripe`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Stripe-Signature': req.headers.get('stripe-signature') ?? '',
    },
    body,
  })

  return new NextResponse(response.body, {
    status: response.status,
    headers: { 'Content-Type': 'application/json' },
  })
}
