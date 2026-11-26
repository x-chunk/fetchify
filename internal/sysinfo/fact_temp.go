package sysinfo

import (
	"context"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// Facts about temperatures.
const (
	// FactCPUTemp is the temperature of the processor.
	FactCPUTemp Fact = "cpu_temp"
	// FactGPUTemp is the temperature of every GPU that reports one.
	FactGPUTemp Fact = "gpu_temp"
)

// cpuSensors are hwmon drivers of processor sensors, in order of
// preference.
var cpuSensors = []string{"k10temp", "coretemp", "zenpower"}

// cpuThermalZones are thermal zone types that measure the processor.
var cpuThermalZones = map[string]bool{"x86_pkg_temp": true, "cpu-thermal": true}

func init() {
	collectors[FactCPUTemp] = fact(collectCPUTemp, func(i *Info) *Value[float64] { return &i.CPUTemp })
	collectors[FactGPUTemp] = fact(collectGPUTemp, func(i *Info) *Value[[]float64] { return &i.GPUTemp })
}

// collectCPUTemp reads the first input of a known hwmon driver, and
// otherwise a processor thermal zone, in degrees Celsius. The fact is
// Absent when the machine has no known sensor.
func collectCPUTemp(_ context.Context, s *session) Value[float64] {
	hwmons, _ := fs.ReadDir(s.fsys(), "sys/class/hwmon")
	for _, driver := range cpuSensors {
		for _, e := range hwmons {
			dir := path.Join("sys/class/hwmon", e.Name())
			if name, _ := readTrimmed(s.fsys(), path.Join(dir, "name")); name != driver {
				continue
			}
			if c, ok := readMilliCelsius(s.fsys(), path.Join(dir, "temp1_input")); ok {
				return present(c)
			}
		}
	}
	zones, _ := fs.ReadDir(s.fsys(), "sys/class/thermal")
	for _, e := range zones {
		if !strings.HasPrefix(e.Name(), "thermal_zone") {
			continue
		}
		dir := path.Join("sys/class/thermal", e.Name())
		if kind, _ := readTrimmed(s.fsys(), path.Join(dir, "type")); !cpuThermalZones[kind] {
			continue
		}
		if c, ok := readMilliCelsius(s.fsys(), path.Join(dir, "temp")); ok {
			return present(c)
		}
	}
	return absent[float64]()
}

// collectGPUTemp reads the hwmon sensor of every GPU, in bus order. The
// NVIDIA driver has no hwmon sensor, so NVIDIA GPUs are asked through
// nvidia-smi, whose lines follow the same bus order. GPUs without a
// reading are left out. The fact is Absent without GPUs and Unknown when
// none of them reports a temperature.
func collectGPUTemp(ctx context.Context, s *session) Value[[]float64] {
	devices := drmDevices(s.fsys())
	if len(devices) == 0 {
		return absent[[]float64]()
	}
	var nvidia []float64
	var asked bool
	var temps []float64
	for _, d := range devices {
		if c, ok := hwmonTemp(s.fsys(), d.dir); ok {
			temps = append(temps, c)
			continue
		}
		if d.vendor != "10de" {
			continue
		}
		if !asked {
			asked = true
			nvidia = nvidiaSMI(ctx, s, "temperature.gpu")
		}
		if len(nvidia) > 0 {
			temps = append(temps, nvidia[0])
			nvidia = nvidia[1:]
		}
	}
	if len(temps) == 0 {
		return Value[[]float64]{}
	}
	return present(temps)
}

// hwmonTemp reads the first temperature input under a device's hwmon
// directory.
func hwmonTemp(fsys fs.FS, device string) (float64, bool) {
	entries, _ := fs.ReadDir(fsys, path.Join(device, "hwmon"))
	for _, e := range entries {
		if c, ok := readMilliCelsius(fsys, path.Join(device, "hwmon", e.Name(), "temp1_input")); ok {
			return c, true
		}
	}
	return 0, false
}

// readMilliCelsius reads a sysfs temperature in millidegrees Celsius.
func readMilliCelsius(fsys fs.FS, name string) (float64, bool) {
	n, ok := readInt(fsys, name)
	return float64(n) / 1000, ok
}

// nvidiaSMI asks nvidia-smi for one numeric property of every NVIDIA GPU
// and returns the values in its order. It returns nil when the tool is
// missing or fails.
func nvidiaSMI(ctx context.Context, s *session, property string) []float64 {
	out, err := s.run(ctx, "nvidia-smi", "--query-gpu="+property, "--format=csv,noheader,nounits")
	if err != nil {
		return nil
	}
	var values []float64
	for line := range strings.SplitSeq(string(out), "\n") {
		if v, err := strconv.ParseFloat(strings.TrimSpace(line), 64); err == nil {
			values = append(values, v)
		}
	}
	return values
}
