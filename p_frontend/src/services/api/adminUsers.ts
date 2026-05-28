import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'
import { adminUserKeys, type AdminUserListParams } from './queryKeys'
import type { AdminUser, AdminUserListResponse } from '../types/adminUser'
import type { components } from '../types/openapi.generated'
import {
  createAdminUserSchema,
  updateAdminUserSchema,
  type CreateAdminUserInput,
  type UpdateAdminUserInput,
} from '../schemas/adminUser'
import { toAdminUser, toAdminUserListResponse } from './adapters/adminUser'

const mockUsers: AdminUser[] = [
  { id: 'u_1001', name: 'Alice Chen', role: 'Super Admin', status: 'active' },
  { id: 'u_1002', name: 'Bob Wang', role: 'Operator', status: 'active' },
  { id: 'u_1003', name: 'Carla Li', role: 'Auditor', status: 'disabled' },
]

export async function listAdminUsers(params: AdminUserListParams): Promise<AdminUserListResponse> {
  try {
    const response = await apiClient.get<components['schemas']['AdminUserListResponse']>(
      '/admin/users',
      { params },
    )

    return toAdminUserListResponse(response.data)
  } catch {
    const start = (params.page - 1) * params.pageSize
    const end = start + params.pageSize

    return {
      items: mockUsers.slice(start, end),
      total: mockUsers.length,
    }
  }
}

export async function getAdminUserDetail(id: string): Promise<AdminUser> {
  try {
    const response = await apiClient.get<components['schemas']['AdminUser']>(
      `/admin/users/${id}`,
    )
    return toAdminUser(response.data)
  } catch {
    const found = mockUsers.find((item) => item.id === id)

    if (!found) {
      throw new Error(`管理员 ${id} 不存在`)
    }

    return found
  }
}

export async function createAdminUser(payload: CreateAdminUserInput): Promise<AdminUser> {
  const validatedPayload = createAdminUserSchema.parse(payload)

  try {
    const response = await apiClient.post<components['schemas']['AdminUser']>(
      '/admin/users',
      validatedPayload,
    )

    return toAdminUser(response.data)
  } catch {
    const created: AdminUser = {
      id: `u_${Date.now()}`,
      ...validatedPayload,
    }
    mockUsers.unshift(created)
    return created
  }
}

export async function updateAdminUser(
  id: string,
  payload: UpdateAdminUserInput,
): Promise<AdminUser> {
  const validatedPayload = updateAdminUserSchema.parse(payload)

  try {
    const response = await apiClient.put<components['schemas']['AdminUser']>(
      `/admin/users/${id}`,
      validatedPayload,
    )
    return toAdminUser(response.data)
  } catch {
    const existing = mockUsers.find((user) => user.id === id)
    if (!existing) {
      throw new Error(`管理员 ${id} 不存在`)
    }

    const updated = { ...existing, ...validatedPayload }
    const index = mockUsers.findIndex((user) => user.id === id)
    mockUsers[index] = updated
    return updated
  }
}

export function useAdminUsersQuery(params: AdminUserListParams) {
  return useQuery({
    queryKey: adminUserKeys.list(params),
    queryFn: () => listAdminUsers(params),
  })
}

export function useAdminUserDetailQuery(id: string) {
  return useQuery({
    queryKey: adminUserKeys.detail(id),
    queryFn: () => getAdminUserDetail(id),
  })
}

export function useCreateAdminUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createAdminUser,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: adminUserKeys.all })
    },
  })
}

export function useUpdateAdminUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateAdminUserInput }) =>
      updateAdminUser(id, payload),
    onSuccess: (_, variables) => {
      void queryClient.invalidateQueries({ queryKey: adminUserKeys.all })
      void queryClient.invalidateQueries({
        queryKey: adminUserKeys.detail(variables.id),
      })
    },
  })
}
