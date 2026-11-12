package authorization

type Permission string

const (
	PermissionStylesView   Permission = "styles.view"
	PermissionStylesCreate Permission = "styles.create"
	PermissionStylesUpdate Permission = "styles.update"
	PermissionStylesDelete Permission = "styles.delete"

	PermissionWorkersView   Permission = "workers.view"
	PermissionWorkersCreate Permission = "workers.create"
	PermissionWorkersUpdate Permission = "workers.update"
	PermissionWorkersDelete Permission = "workers.delete"

	PermissionBundlesView   Permission = "bundles.view"
	PermissionBundlesCreate Permission = "bundles.create"
	PermissionBundlesUpdate Permission = "bundles.update"
	PermissionBundlesDelete Permission = "bundles.delete"

	PermissionProductionView   Permission = "production.view"
	PermissionProductionCreate Permission = "production.create"
	PermissionProductionUpdate Permission = "production.update"
	PermissionProductionDelete Permission = "production.delete"

	PermissionOrganizationView   Permission = "organization.view"
	PermissionOrganizationUpdate Permission = "organization.update"

	PermissionUsersView   Permission = "users.view"
	PermissionUsersManage Permission = "users.manage"

	PermissionDashboardView Permission = "dashboard.view"
)
