import { z } from 'zod'

export const loginRequestSchema = z.object({
  username: z.string().min(2, '用户名至少 2 个字符'),
  password: z.string().min(6, '密码至少 6 个字符'),
})

export type LoginRequestInput = z.input<typeof loginRequestSchema>
export type LoginRequestPayload = z.output<typeof loginRequestSchema>
