// Package distro knows Linux distributions: it parses os-release and
// holds the table of distributions with their names, colors and logos.
//
// The package is pure: it reads nothing from the machine. sysinfo
// collects the content of os-release and ui passes it here.
package distro
