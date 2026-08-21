import { z } from 'zod'

export const organizationUserProfileSchema = z.object({
  displayName: z.string().trim().min(1).max(64),
  avatar: z.string().trim().max(255),
  email: z.string().trim().max(128),
  phone: z.string().trim().max(32),
})

export const organizationUserPositionSchema = z.object({
  departmentId: z.number().int().nonnegative(),
  positionId: z.number().int().nonnegative(),
})

export const organizationUserStatusSchema = z.union([z.literal(0), z.literal(1), z.literal(2)])

export const organizationPositionRolesSchema = z.array(z.number().int().positive()).max(100)
