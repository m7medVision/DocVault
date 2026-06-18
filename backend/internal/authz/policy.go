package authz

const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
	RoleOwner  = "owner"
)

const (
	ResourceDocuments     = "documents"
	ResourceFolders       = "folders"
	ResourceTags          = "tags"
	ResourceSearch        = "search"
	ResourceReminders     = "reminders"
	ResourceNotifications = "notifications"
	ResourceAdminAudit    = "admin.audit"
	ResourceAdminMembers  = "admin.members"
	ResourceAdminPolicies = "admin.policies"
	ResourceACL           = "admin.acl"
	ResourceGroups        = "groups"
)

const (
	ActionRead       = "read"
	ActionWrite      = "write"
	ActionDelete     = "delete"
	ActionInvite     = "invite"
	ActionUpdateRole = "update_role"
)

type Permission struct {
	Role   string
	Object string
	Action string
}

var defaultHierarchy = [][2]string{
	{RoleOwner, RoleAdmin},
	{RoleAdmin, RoleMember},
	{RoleMember, RoleViewer},
}

var defaultPermissions = []Permission{
	{RoleViewer, ResourceDocuments, ActionRead},
	{RoleViewer, ResourceFolders, ActionRead},
	{RoleViewer, ResourceTags, ActionRead},
	{RoleViewer, ResourceSearch, ActionRead},
	{RoleViewer, ResourceReminders, ActionRead},
	{RoleViewer, ResourceNotifications, ActionRead},

	{RoleMember, ResourceDocuments, ActionWrite},
	{RoleMember, ResourceFolders, ActionWrite},
	{RoleMember, ResourceReminders, ActionWrite},
	{RoleMember, ResourceNotifications, ActionWrite},

	{RoleAdmin, ResourceDocuments, ActionDelete},
	{RoleAdmin, ResourceAdminAudit, ActionRead},
	{RoleAdmin, ResourceAdminMembers, ActionRead},
	{RoleAdmin, ResourceAdminMembers, ActionInvite},
	{RoleAdmin, ResourceAdminPolicies, ActionRead},
	{RoleAdmin, ResourceAdminPolicies, ActionWrite},
	{RoleAdmin, ResourceAdminPolicies, ActionDelete},
	{RoleAdmin, ResourceACL, ActionRead},
	{RoleAdmin, ResourceACL, ActionWrite},
	{RoleAdmin, ResourceACL, ActionDelete},
	{RoleAdmin, ResourceGroups, ActionRead},
	{RoleAdmin, ResourceGroups, ActionWrite},
	{RoleAdmin, ResourceGroups, ActionDelete},

	{RoleOwner, ResourceAdminMembers, ActionUpdateRole},
}
