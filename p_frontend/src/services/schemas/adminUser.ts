import { z } from 'zod'

export const adminUserRoleSchema = z.enum(['Super Admin', 'Operator', 'Auditor'])
export const adminUserStatusSchema = z.enum(['active', 'disabled'])

export const createAdminUserSchema = z.object({
  name: z.string().min(2, '姓名至少 2 个字符').max(50, '姓名最多 50 个字符'),
  role: adminUserRoleSchema,
  status: adminUserStatusSchema,
})

export const updateAdminUserSchema = createAdminUserSchema

export type CreateAdminUserInput = z.input<typeof createAdminUserSchema>
export type CreateAdminUserPayload = z.output<typeof createAdminUserSchema>
export type UpdateAdminUserInput = z.input<typeof updateAdminUserSchema>
export type UpdateAdminUserPayload = z.output<typeof updateAdminUserSchema>
