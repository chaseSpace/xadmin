package consts

const (
	PermissionStatusEnabled  int32 = 1
	PermissionStatusDisabled int32 = 0
)

const (
	PermissionMenuTypeDirectory int32 = 1
	PermissionMenuTypeMenu      int32 = 2
	PermissionMenuTypeButton    int32 = 3
)

const (
	PermissionRoleTypeSystem int32 = 1
	PermissionRoleTypeCustom int32 = 2
)

const PermissionRoleCodeSuperAdmin = "super_admin"

const (
	PermissionPositionsEditProfile = "organization.positions.edit_profile"
	PermissionPositionsAssignRoles = "organization.positions.assign_roles"
	PermissionUsersEditProfile     = "organization.users.edit_profile"
	PermissionUsersAssignPosition  = "organization.users.assign_position"
	PermissionUsersChangeStatus    = "organization.users.change_status"
	PermissionUsersResetPassword   = "organization.users.reset_password"
	PermissionRolesEditProfile     = "permission.roles.edit_profile"
	PermissionRolesAssignMenus     = "permission.roles.assign_menus"
	PermissionMenusManageSchema    = "permission.menus.manage_schema"
)
