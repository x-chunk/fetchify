// Package sysinfo collects facts about the Linux machine fetchify runs on.
//
// A fact is one independently collectable piece of data, such as the
// kernel release or the list of disks. Collect gathers the requested
// facts concurrently, each with its own time limit, and returns an Info
// with a Value per fact. A Value is Present, Absent when the fact does
// not exist on the machine, or Unknown when it was not requested or could
// not be determined. Collection never fails.
//
// The package returns plain data only: no formatting, no colors and no
// fallback strings. Collectors read the machine only through a Source, so
// tests run them against fixture trees.
package sysinfo
