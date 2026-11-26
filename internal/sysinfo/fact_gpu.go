package sysinfo

import (
	"bufio"
	"bytes"
	"context"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

// FactGPU is the list of display adapters.
const FactGPU Fact = "gpu"

func init() {
	collectors[FactGPU] = fact(collectGPU, func(i *Info) *Value[[]string] { return &i.GPU })
}

// pciDevice is a display adapter found in sysfs. IDs are lower-case hex
// without the 0x prefix; dir is its device directory in sysfs.
type pciDevice struct {
	addr, vendor, device string
	dir                  string
}

// displayClasses are the PCI class codes of display controllers: VGA, 3D
// and other display controllers. Discrete GPUs of laptops are usually 3D
// controllers.
var displayClasses = []string{"0300", "0302", "0380"}

// pciIDsPaths are the locations of the PCI ID database, in order.
var pciIDsPaths = []string{
	"usr/share/hwdata/pci.ids",
	"usr/share/misc/pci.ids",
	"usr/share/pci.ids",
}

// wellKnownVendors names the GPU vendors when no database is available.
var wellKnownVendors = map[string]string{
	"10de": "NVIDIA",
	"1002": "AMD",
	"8086": "Intel",
}

// cardName matches the DRM card directories, as opposed to connectors
// such as "card0-HDMI-A-1" and render nodes.
var cardName = regexp.MustCompile(`^card[0-9]+$`)

// collectGPU lists the display adapters in bus order. Names come from the
// PCI ID database, then from lspci, then from the well-known vendor IDs.
// The fact is Absent when there are no display adapters.
func collectGPU(ctx context.Context, s *session) Value[[]string] {
	devices := drmDevices(s.fsys())
	if len(devices) == 0 {
		return absent[[]string]()
	}
	names := make([]string, len(devices))
	if db, ok := readPCIIDs(s.fsys()); ok {
		ids := lookupPCIIDs(db, devices)
		for i, d := range devices {
			names[i] = ids[d.vendor+":"+d.device]
		}
	} else if out, err := s.run(ctx, "lspci", "-mm"); err == nil {
		slots := parseLspciMM(string(out))
		for i, d := range devices {
			names[i] = slots[strings.TrimPrefix(d.addr, "0000:")]
		}
	}
	for i, d := range devices {
		if names[i] == "" {
			vendor := wellKnownVendors[d.vendor]
			if vendor == "" {
				vendor = d.vendor
			}
			names[i] = vendor + " " + d.device
		}
	}
	return present(names)
}

// drmDevices returns the PCI display adapters behind the DRM cards,
// deduplicated by PCI address and sorted by it.
func drmDevices(fsys fs.FS) []pciDevice {
	entries, err := fs.ReadDir(fsys, "sys/class/drm")
	if err != nil {
		return nil
	}
	var devices []pciDevice
	for _, e := range entries {
		// sysfs entries are symbolic links, so IsDir is not checked.
		if !cardName.MatchString(e.Name()) {
			continue
		}
		dir := path.Join("sys/class/drm", e.Name(), "device")
		class, err := readHexID(fsys, path.Join(dir, "class"))
		if err != nil || !slices.ContainsFunc(displayClasses, func(c string) bool { return strings.HasPrefix(class, c) }) {
			continue
		}
		vendor, err1 := readHexID(fsys, path.Join(dir, "vendor"))
		device, err2 := readHexID(fsys, path.Join(dir, "device"))
		if err1 != nil || err2 != nil {
			continue
		}
		addr := pciSlot(fsys, dir)
		if addr == "" {
			addr = e.Name()
		}
		if slices.ContainsFunc(devices, func(d pciDevice) bool { return d.addr == addr }) {
			continue
		}
		devices = append(devices, pciDevice{addr: addr, vendor: vendor, device: device, dir: dir})
	}
	slices.SortFunc(devices, func(a, b pciDevice) int { return strings.Compare(a.addr, b.addr) })
	return devices
}

// readHexID reads a sysfs ID such as "0x10de" and returns it in lower
// case without the prefix.
func readHexID(fsys fs.FS, name string) (string, error) {
	s, err := readTrimmed(fsys, name)
	return strings.ToLower(strings.TrimPrefix(s, "0x")), err
}

// pciSlot returns the PCI address of a device from its uevent file, or
// "" when it is not known.
func pciSlot(fsys fs.FS, dir string) string {
	data, err := fs.ReadFile(fsys, path.Join(dir, "uevent"))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if slot, found := strings.CutPrefix(line, "PCI_SLOT_NAME="); found {
			return strings.TrimSpace(slot)
		}
	}
	return ""
}

// readPCIIDs returns the content of the first PCI ID database that
// exists.
func readPCIIDs(fsys fs.FS) ([]byte, bool) {
	for _, name := range pciIDsPaths {
		if data, err := fs.ReadFile(fsys, name); err == nil {
			return data, true
		}
	}
	return nil, false
}

// lookupPCIIDs resolves the devices in a pci.ids database to
// "<vendor name> <device name>", keyed by "vendor:device". A device whose
// vendor is known but the device is not gets the vendor name and its ID.
//
// The database lists vendors without indentation and their devices
// indented by one tab; deeper lines are subsystems. Only the vendors of
// the devices are parsed, which keeps the cost low for a file of more
// than a megabyte.
func lookupPCIIDs(db []byte, devices []pciDevice) map[string]string {
	wanted := map[string][]string{} // vendor -> device IDs
	for _, d := range devices {
		wanted[d.vendor] = append(wanted[d.vendor], d.device)
	}
	names := map[string]string{}
	var vendor, vendorName string
	scanner := bufio.NewScanner(bytes.NewReader(db))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "" || line[0] == '#':
		case line[0] != '\t':
			vendor, vendorName = "", ""
			id, name, ok := splitPCIIDLine(line)
			if ok && wanted[id] != nil {
				vendor, vendorName = id, name
				for _, device := range wanted[id] {
					names[id+":"+device] = name + " " + device
				}
			}
		case vendor != "" && !strings.HasPrefix(line, "\t\t"):
			id, name, ok := splitPCIIDLine(line[1:])
			if ok && slices.Contains(wanted[vendor], id) {
				names[vendor+":"+id] = vendorName + " " + name
			}
		}
	}
	return names
}

// splitPCIIDLine splits "10de  NVIDIA Corporation" into the ID and the
// name.
func splitPCIIDLine(line string) (id, name string, ok bool) {
	id, name, ok = strings.Cut(line, "  ")
	return strings.ToLower(id), strings.TrimSpace(name), ok && len(id) == 4
}

// parseLspciMM maps the slots in `lspci -mm` output to
// "<vendor name> <device name>". A line looks like
// `01:00.0 "VGA compatible controller" "NVIDIA Corporation" "GA104" ...`.
func parseLspciMM(out string) map[string]string {
	names := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		slot, rest, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		fields := quotedFields(rest)
		if len(fields) >= 3 {
			names[slot] = strings.TrimSpace(fields[1] + " " + fields[2])
		}
	}
	return names
}

// quotedFields returns the double-quoted strings of s in order. Text
// outside quotes, such as lspci's "-ra1" revision flags, is skipped.
func quotedFields(s string) []string {
	var fields []string
	for {
		start := strings.IndexByte(s, '"')
		if start < 0 {
			return fields
		}
		end := strings.IndexByte(s[start+1:], '"')
		if end < 0 {
			return fields
		}
		fields = append(fields, s[start+1:start+1+end])
		s = s[start+end+2:]
	}
}
