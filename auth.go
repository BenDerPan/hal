package hal

import (
	"fmt"
	"os"
	"strings"
)

var admins []string

// NewAuth returns a pointer to an initialized Auth
func NewAuth(r *Robot) *Auth {
	a := &Auth{robot: r}
	if admins := os.Getenv("HAL_AUTH_ADMIN"); admins != "" {
		a.admins = strings.Split(admins, ",")
	}

	r.Handle(assignUserRoleHandler)

	return a
}

// Auth type to group authentication methods
type Auth struct {
	robot  *Robot
	admins []string
}

// HasRole checks whether a user located by id has a given role(s)
func (a *Auth) HasRole(id string, roles ...string) bool {
	u, err := a.robot.Users.Get(id)
	if err != nil {
		return false
	}

	if len(u.Roles) == 0 {
		return false
	}

	for _, r := range roles {
		for _, b := range u.Roles {
			if b == r {
				return true
			}
		}
	}

	return false
}

// UsersWithRole returns a slice of Users that have a given role
func (a *Auth) UsersWithRole(role string) (users []User) {
	for _, u := range a.robot.Users.All() {
		if a.HasRole(u.ID, role) {
			users = append(users, u)
		}
	}
	return
}

func (a *Auth) addRole(u User, r string) error {
	if r == "admin" {
		return fmt.Errorf(`the "admin" role can only be defined by the HAL_AUTH_ADMIN environment variable`)
	}

	if a.HasRole(u.ID, r) {
		return fmt.Errorf("%s already has the %s role", u.Name, r)
	}

	u.Roles = append(u.Roles, r)
	a.robot.Users.Set(u.ID, u)

	return nil
}

func assignUserRole(res *Response) error {
	n := strings.TrimSpace(res.Match[1])
	r := strings.ToLower(res.Match[3])

	for _, i := range []string{"", "who", "what", "where", "when", "why"} {
		if i == n {
			return nil // don't match
		}
	}

	u, err := res.Robot.Users.GetByName(n)
	if err != nil {
		return res.Reply(n + " does not exist")
	}

	if err := res.Robot.Auth.addRole(u, r); err != nil {
		return res.Reply(err.Error())
	}

	return res.Reply(fmt.Sprintf("%s now has the %s role", n, r))
}

var assignUserRoleHandler = &Handler{
	Pattern: `(?i)@?(.+) (has) (["'\w: -_]+) (role)`,
	Method:  RESPOND,
	Run:     assignUserRole,
}
