// Stripe client-side utilities

export function getStripePublishableKey(): string {
  const key = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  if (!key) throw new Error('NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY is not set')
  return key
}

// Redirect to Stripe Checkout for session payment.
// The checkout_url comes from POST /v1/sessions.
export function redirectToCheckout(checkoutUrl: string) {
  window.location.href = checkoutUrl
}
