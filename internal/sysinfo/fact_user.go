package sysinfo

import (
	"context"
	"strconv"
)

// User is the identity the process runs as.
type User struct {
	// Name is the login name.
	Name string
	// UID is the effective user ID; 0 is root.
	UID int
}

// FactUser is the user the process runs as.
const FactUser Fact = "user"

func init() {
	collectors[FactUser] = fact(collectUser, func(i *Info) *Value[User] { return &i.User })
}

// collectUser looks up the name of the effective user. When the user
// database has no entry, which happens in containers with an arbitrary
// UID, it falls back to $USER and then to the UID as text. The fact is
// therefore always present.
func collectUser(_ context.Context, s *session) Value[User] {
	u := User{UID: s.src.UID}
	if s.src.LookupUser != nil {
		if name, err := s.src.LookupUser(u.UID); err == nil {
			u.Name = name
		}
	}
	if u.Name == "" {
		u.Name = s.getenv("USER")
	}
	if u.Name == "" {
		u.Name = strconv.Itoa(u.UID)
	}
	return present(u)
}
