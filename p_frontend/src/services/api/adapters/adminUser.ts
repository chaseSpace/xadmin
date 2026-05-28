import type { components } from '../../types/openapi.generated'
import type { AdminUser, AdminUserListResponse } from '../../types/adminUser'

type ApiAdminUser = components['schemas']['AdminUser']
type ApiAdminUserListResponse = components['schemas']['AdminUserListResponse']

export function toAdminUser(apiUser: ApiAdminUser): AdminUser {
  return {
    id: apiUser.id,
    name: apiUser.name,
    role: apiUser.role,
    status: apiUser.status,
  }
}

export function toAdminUserListResponse(
  apiResponse: ApiAdminUserListResponse,
): AdminUserListResponse {
  return {
    items: apiResponse.items.map(toAdminUser),
    total: apiResponse.total,
  }
}
