import StreamViewer from '@/components/session/StreamViewer'
import BillingTimer from '@/components/session/BillingTimer'

interface Props {
  params: { id: string }
}

// This page is a Client Component boundary — WebRTC requires browser APIs.
// The page itself can be a Server Component; only StreamViewer is a Client Component.
export default async function SessionPage({ params }: Props) {
  // TODO: Validate session ownership (auth check on server)
  // TODO: Fetch session status to get signaling_url

  const signalingUrl = `${process.env.NEXT_PUBLIC_WS_URL}/v1/signal/${params.id}`

  return (
    <div className="flex flex-col h-screen bg-black">
      <div className="flex-1 relative">
        <StreamViewer
          sessionId={params.id}
          signalingUrl={signalingUrl}
        />
      </div>
      <div className="h-12 bg-gray-900 border-t border-gray-800 flex items-center px-4">
        <BillingTimer sessionId={params.id} />
      </div>
    </div>
  )
}
