// Path: ./blogX_server/models/enum/role_type.go

package enum

type RoleType uint8

const (
	AdminRoleType RoleType = 1
	UserRoleType  RoleType = 2
	GuestRoleType RoleType = 3
)
