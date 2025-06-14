// Path: ./models/enum/role_type.go

package enum

type RoleType uint8

const (
	AdminRoleType RoleType = 1
	UserRoleType  RoleType = 2
	GuestRoleType RoleType = 3
)

func (r RoleType) String() string {
	switch r {
	case AdminRoleType:
		return "Admin"
	case UserRoleType:
		return "User"
	case GuestRoleType:
		return "Guest"
	}
	return ""
}
