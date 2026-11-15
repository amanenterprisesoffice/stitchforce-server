package authorization

type Permission string

const (
	PermissionStylesView   Permission = "styles.view"
	PermissionStylesCreate Permission = "styles.create"
	PermissionStylesUpdate Permission = "styles.update"
	PermissionStylesDelete Permission = "styles.delete"

	PermissionSizesView   Permission = "sizes.view"
	PermissionSizesCreate Permission = "sizes.create"
	PermissionSizesUpdate Permission = "sizes.update"
	PermissionSizesDelete Permission = "sizes.delete"

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

	PermissionPaymentsView   Permission = "payments.view"
	PermissionPaymentsCreate Permission = "payments.create"
	PermissionPaymentsUpdate Permission = "payments.update"

	PermissionOrganizationView   Permission = "organization.view"
	PermissionOrganizationUpdate Permission = "organization.update"

	PermissionUsersView   Permission = "users.view"
	PermissionUsersManage Permission = "users.manage"

	PermissionDashboardView Permission = "dashboard.view"
)

var PermissionMessages = map[Permission]string{
	PermissionStylesView:   "You don't have access to view styles.",
	PermissionStylesCreate: "You don't have access to create styles.",
	PermissionStylesUpdate: "You don't have access to update styles.",
	PermissionStylesDelete: "You don't have access to delete styles.",

	PermissionSizesView:   "You don't have access to view sizes.",
	PermissionSizesCreate: "You don't have access to create sizes.",
	PermissionSizesUpdate: "You don't have access to update sizes.",
	PermissionSizesDelete: "You don't have access to delete sizes.",

	PermissionWorkersView:   "You don't have access to view garment workers.",
	PermissionWorkersCreate: "You don't have access to create garment workers.",
	PermissionWorkersUpdate: "You don't have access to update garment workers.",
	PermissionWorkersDelete: "You don't have access to delete garment workers.",

	PermissionBundlesView:   "You don't have access to view bundles.",
	PermissionBundlesCreate: "You don't have access to create bundles.",
	PermissionBundlesUpdate: "You don't have access to update bundles.",
	PermissionBundlesDelete: "You don't have access to delete bundles.",

	PermissionProductionView:   "You don't have access to view production records.",
	PermissionProductionCreate: "You don't have access to create production records.",
	PermissionProductionUpdate: "You don't have access to update production records.",
	PermissionProductionDelete: "You don't have access to delete production records.",

	PermissionPaymentsView:   "You don't have access to view payments.",
	PermissionPaymentsCreate: "You don't have access to create payments.",
	PermissionPaymentsUpdate: "You don't have access to update payments.",

	PermissionOrganizationView:   "You don't have access to view organization details.",
	PermissionOrganizationUpdate: "You don't have access to update organization details.",

	PermissionUsersView:   "You don't have access to view users.",
	PermissionUsersManage: "You don't have access to manage users.",

	PermissionDashboardView: "You don't have access to view the dashboard.",
}
