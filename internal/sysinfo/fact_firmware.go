package sysinfo

import (
	"context"
	"path"
	"strings"
)

// Facts read from the DMI tables besides the host model.
const (
	// FactBoard is the vendor and the name of the main board.
	FactBoard Fact = "board"
	// FactBIOS is the firmware vendor, version and release date.
	FactBIOS Fact = "bios"
)

// dmiDir is the sysfs directory of the DMI identification strings.
const dmiDir = "sys/devices/virtual/dmi/id"

// BIOS identifies the firmware. Placeholder values are empty.
type BIOS struct {
	// Vendor is the firmware vendor, for example "American Megatrends".
	Vendor string
	// Version is the firmware version.
	Version string
	// Date is the release date as the firmware reports it, usually
	// "MM/DD/YYYY".
	Date string
}

func init() {
	collectors[FactBoard] = fact(collectBoard, func(i *Info) *Value[string] { return &i.Board })
	collectors[FactBIOS] = fact(collectBIOS, func(i *Info) *Value[BIOS] { return &i.BIOS })
}

// readDMI reads a DMI field and returns "" for placeholders.
func readDMI(s *session, field string) string {
	value, _ := readTrimmed(s.fsys(), path.Join(dmiDir, field))
	if isPlaceholder(value) {
		return ""
	}
	return value
}

// collectBoard joins the board vendor and name. The fact is Unknown when
// both are missing or placeholders.
func collectBoard(_ context.Context, s *session) Value[string] {
	board := strings.TrimSpace(readDMI(s, "board_vendor") + " " + readDMI(s, "board_name"))
	if board == "" {
		return Value[string]{}
	}
	return present(board)
}

// collectBIOS reads the firmware identification. The fact is Unknown when
// every field is missing or a placeholder.
func collectBIOS(_ context.Context, s *session) Value[BIOS] {
	b := BIOS{
		Vendor:  readDMI(s, "bios_vendor"),
		Version: readDMI(s, "bios_version"),
		Date:    readDMI(s, "bios_date"),
	}
	if b == (BIOS{}) {
		return Value[BIOS]{}
	}
	return present(b)
}
