package sysinfo

import "testing"

func TestUser(t *testing.T) {
	tests := []struct {
		name    string
		machine machine
		want    Value[User]
	}{
		{"user database", ubuntuDesktop, present(User{Name: "alice", UID: 1000})},
		{"root", sshServer, present(User{Name: "root", UID: 0})},
		{"unknown UID", container, present(User{Name: "100123", UID: 100123})},
		{"from $USER", machine{uid: 5, env: map[string]string{"USER": "eve"}}, present(User{Name: "eve", UID: 5})},
		{"no lookup", machine{}, present(User{Name: "0"})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOne(t, tt.machine, FactUser).User; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
