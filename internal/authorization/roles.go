package authorization

import "stitchforce-server/internal/entity"

var RolePermissions = map[entity.UserRole]map[Permission]bool{
	entity.UserRoleOrgAdmin: {
		PermissionStylesView:   true,
		PermissionStylesCreate: true,
		PermissionStylesUpdate: true,
		PermissionStylesDelete: true,

		PermissionWorkersView:   true,
		PermissionWorkersCreate: true,
		PermissionWorkersUpdate: true,
		PermissionWorkersDelete: true,

		PermissionBundlesView:   true,
		PermissionBundlesCreate: true,
		PermissionBundlesUpdate: true,
		PermissionBundlesDelete: true,

		PermissionProductionView:   true,
		PermissionProductionCreate: true,
		PermissionProductionUpdate: true,
		PermissionProductionDelete: true,

		PermissionOrganizationView:   true,
		PermissionOrganizationUpdate: true,

		PermissionUsersView:   true,
		PermissionUsersManage: true,

		PermissionDashboardView: true,
	},

	entity.UserRoleAdmin: {
		PermissionStylesView:   true,
		PermissionStylesCreate: true,
		PermissionStylesUpdate: true,
		PermissionStylesDelete: true,

		PermissionWorkersView:   true,
		PermissionWorkersCreate: true,
		PermissionWorkersUpdate: true,
		PermissionWorkersDelete: true,

		PermissionBundlesView:   true,
		PermissionBundlesCreate: true,
		PermissionBundlesUpdate: true,
		PermissionBundlesDelete: true,

		PermissionProductionView:   true,
		PermissionProductionCreate: true,
		PermissionProductionUpdate: true,
		PermissionProductionDelete: true,

		PermissionOrganizationView: true,

		PermissionUsersView:   true,
		PermissionUsersManage: true,

		PermissionDashboardView: true,
	},

	entity.UserRoleStaff: {
		PermissionStylesView:       true,
		PermissionWorkersView:      true,
		PermissionBundlesView:      true,
		PermissionOrganizationView: true,

		PermissionProductionView:   true,
		PermissionProductionCreate: true,

		PermissionDashboardView: true,
	},
}

func HasPermission(roleValue any, permission Permission) bool {
	var role entity.UserRole

	switch value := roleValue.(type) {
	case entity.UserRole:
		role = value
	case string:
		role = entity.UserRole(value)
	default:
		return false
	}

	if !role.IsValid() {
		return false
	}

	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	return permissions[permission]
}
