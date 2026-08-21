import type { AuthUserProfile } from '../store/auth'

export const permissionKeys = {
  positionsEditProfile: 'organization.positions.edit_profile',
  positionsAssignRoles: 'organization.positions.assign_roles',
  usersEditProfile: 'organization.users.edit_profile',
  usersAssignPosition: 'organization.users.assign_position',
  usersChangeStatus: 'organization.users.change_status',
  usersResetPassword: 'organization.users.reset_password',
  usersDelete: 'organization.users.delete',
  positionsDelete: 'organization.positions.delete',
  rolesEditProfile: 'permission.roles.edit_profile',
  rolesAssignMenus: 'permission.roles.assign_menus',
  rolesDelete: 'permission.roles.delete',
  menusManageSchema: 'permission.menus.manage_schema',
} as const

export function hasPermission(
  user: AuthUserProfile | null | undefined,
  permissionKey: string,
): boolean {
  return Boolean(user?.permissionKeys.includes(permissionKey))
}
