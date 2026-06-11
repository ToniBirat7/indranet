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
  email: z.string().email(),
  password: z.string().min(8),
  name: z.string().min(1).max(100),
})

export const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
})

export const InputEventSchema = z.discriminatedUnion('type', [
  z.object({
    type: z.literal('keyboard'),
    action: z.enum(['down', 'up']),
    key: z.number().int(),
    modifiers: z.number().int().default(0),
  }),
  z.object({
    type: z.literal('mouse'),
    dx: z.number(),
    dy: z.number(),
    buttons: z.number().int().default(0),
    wheel: z.number().default(0),
  }),
  z.object({
    type: z.literal('gamepad'),
    leftX: z.number().min(-1).max(1),
    leftY: z.number().min(-1).max(1),
    rightX: z.number().min(-1).max(1),
    rightY: z.number().min(-1).max(1),
    buttons: z.number().int(),
    leftTrigger: z.number().min(0).max(1),
    rightTrigger: z.number().min(0).max(1),
  }),
])
