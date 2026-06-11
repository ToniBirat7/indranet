import { z } from 'zod'

export const SessionStateSchema = z.enum([
  'CREATED', 'AUTHORIZED', 'ACTIVE', 'ENDING', 'ENDED', 'FAILED',
])

// IDs use a prefixed hex format: usr_<hex24>, hst_<hex24>, ses_<hex24>
const prefixedId = z.string().regex(/^[a-z]{3}_[0-9a-f]{24}$/)

export const CreateSessionSchema = z.object({
  host_id: prefixedId,
  duration_minutes: z.number().int().min(15).max(480),
})

export const RegisterSchema = z.object({
  email: z.string().email().max(254),
  password: z.string().min(8).max(72),
  name: z.string().min(1).max(80),
})

export const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
})

// InputEventSchema matches the compact wire format sent over RTCDataChannel.
// Keys use single-char shorthand to minimise JSON overhead on the 60fps data path.
export const InputEventSchema = z.discriminatedUnion('t', [
  z.object({ t: z.literal('k'), e: z.enum(['d', 'u']), c: z.string(), k: z.number().int() }),
  z.object({ t: z.literal('m'), dx: z.number(), dy: z.number() }),
  z.object({ t: z.literal('mb'), e: z.enum(['d', 'u']), b: z.number().int() }),
  z.object({ t: z.literal('mw'), d: z.number() }),
  z.object({
    t: z.literal('gp'),
    lx: z.number().min(-1).max(1),
    ly: z.number().min(-1).max(1),
    rx: z.number().min(-1).max(1),
    ry: z.number().min(-1).max(1),
    bt: z.number().int(),
    lt: z.number().min(0).max(1),
    rt: z.number().min(0).max(1),
  }),
])
