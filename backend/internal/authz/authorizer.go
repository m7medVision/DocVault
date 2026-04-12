package authz

type Authorizer interface {
	Enforce(args ...interface{}) (bool, error)
	GetRolesForUser(userID string) ([]string, error)
	GetUsersInRole(role string) ([]string, error)
	AddRoleForUser(userID, role string) error
	DeleteRoleForUser(userID, role string) error
}
