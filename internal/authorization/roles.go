package authorization

import "stitchforce-server/internal/entity"

var RolePermissions = map[entity.UserRole]map[Permission]bool{
	entity.UserRoleOrgAdmin: {
		// Styles
		PermissionStylesView:   true,
		PermissionStylesCreate: true,
		PermissionStylesUpdate: true,
		PermissionStylesDelete: true,

		// Sizes
		PermissionSizesView:   true,
		PermissionSizesCreate: true,
		PermissionSizesUpdate: true,
		PermissionSizesDelete: true,

		// Workers
		PermissionWorkersView:   true,
		PermissionWorkersCreate: true,
		PermissionWorkersUpdate: true,
		PermissionWorkersDelete: true,

		// Bundles
		PermissionBundlesView:   true,
		PermissionBundlesCreate: true,
		PermissionBundlesUpdate: true,
		PermissionBundlesDelete: true,

		// Production
		PermissionProductionView:   true,
		PermissionProductionCreate: true,
		PermissionProductionUpdate: true,
		PermissionProductionDelete: true,

		// Payments
		PermissionPaymentsView:   true,
		PermissionPaymentsCreate: true,
		PermissionPaymentsUpdate: true,

		// Organization
		PermissionOrganizationView:   true,
		PermissionOrganizationUpdate: true,

		// Users
		PermissionUsersView:   true,
		PermissionUsersManage: true,

		// Dashboard
		PermissionDashboardView: true,
	},

	entity.UserRoleAdmin: {
		// Styles
		PermissionStylesView:   true,
		PermissionStylesCreate: true,
		PermissionStylesUpdate: true,
		PermissionStylesDelete: true,

		// Sizes
		PermissionSizesView:   true,
		PermissionSizesCreate: true,
		PermissionSizesUpdate: true,
		PermissionSizesDelete: true,

		// Workers
		PermissionWorkersView:   true,
		PermissionWorkersCreate: true,
		PermissionWorkersUpdate: true,
		PermissionWorkersDelete: true,

		// Bundles
		PermissionBundlesView:   true,
		PermissionBundlesCreate: true,
		PermissionBundlesUpdate: true,
		PermissionBundlesDelete: true,

		// Production
		PermissionProductionView:   true,
		PermissionProductionCreate: true,
		PermissionProductionUpdate: true,
		PermissionProductionDelete: true,

		// Payments
		PermissionPaymentsView:   true,
		PermissionPaymentsCreate: true,
		PermissionPaymentsUpdate: true,

		// Organization
		PermissionOrganizationView: true,

		// Users
		PermissionUsersView:   true,
		PermissionUsersManage: true,

		// Dashboard
		PermissionDashboardView: true,
	},

	entity.UserRoleStaff: {
		// Styles
		PermissionStylesView: true,

		// Workers
		PermissionWorkersView: true,

		// Bundles
		PermissionBundlesView: true,

		// Sizes
		PermissionSizesView: true,

		// Production
		PermissionProductionView:   true,
		PermissionProductionCreate: true,
		PermissionProductionUpdate: true,
		PermissionProductionDelete: true,

		// Payments
		PermissionPaymentsView:   true,
		PermissionPaymentsCreate: true,
		PermissionPaymentsUpdate: true,

		// Organization
		PermissionOrganizationView: true,

		// Users
		PermissionUsersView: true,

		// Dashboard
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

func GetPermissions(roleValue any) []Permission {
	var role entity.UserRole

	switch value := roleValue.(type) {
	case entity.UserRole:
		role = value
	case string:
		role = entity.UserRole(value)
	default:
		return nil
	}

	if !role.IsValid() {
		return nil
	}

	permissions, exists := RolePermissions[role]
	if !exists {
		return nil
	}

	result := make([]Permission, 0, len(permissions))

	for permission, allowed := range permissions {
		if allowed {
			result = append(result, permission)
		}
	}

	return result
}
