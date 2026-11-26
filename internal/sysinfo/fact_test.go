package sysinfo

import "testing"

func TestValueStates(t *testing.T) {
	var zero Value[string]
	if zero.State != Unknown {
		t.Errorf("zero Value state = %v, want unknown", zero.State)
	}
	if v := present("x"); v.State != Present || v.V != "x" {
		t.Errorf("present = %+v", v)
	}
	if v := absent[int](); v.State != Absent || v.V != 0 {
		t.Errorf("absent = %+v", v)
	}
	for s, want := range map[State]string{Unknown: "unknown", Present: "present", Absent: "absent", 9: "unknown"} {
		if got := s.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", s, got, want)
		}
	}
}
