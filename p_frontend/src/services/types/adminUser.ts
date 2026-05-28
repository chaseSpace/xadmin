export type AdminUser = {
  id: string
  name: string
  role: 'Super Admin' | 'Operator' | 'Auditor'
  status: 'active' | 'disabled'
}

export type AdminUserListResponse = {
  items: AdminUser[]
  total: number
}
