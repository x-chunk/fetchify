# Goals

Work that is intentionally left out of fetchify 2.0. `plan.md` covers that
version; nothing in this file is part of it. The design in `plan.md` must
not block any item in "Deferred".

## Deferred

### Other operating systems

macOS and the BSDs. Version 2.0 reads Linux interfaces directly (`/proc`,
`/sys`, `os-release`).

Constraints that keep it possible:

- Collectors read the machine only through `sysinfo.Source`, and the rest
  of the program sees only `sysinfo.Info`. A port adds collectors behind
  build tags and touches nothing above `sysinfo`.
- `Absent` and `Unknown` already express "this fact does not exist here",
  so a platform that lacks a fact needs no special handling in the UI.
- `distro` is keyed by an identifier, not by Linux specifics: other
  systems can get entries and logos in the same table.

### Named themes

Presets such as `theme = "minimal"` that set several configuration keys
at once. Deferred until the configuration schema has settled, so that a
theme is a plain set of existing keys and not a second mechanism.

### Items that run user commands

A configuration entry that shows the output of a command. It runs
arbitrary commands at every shell start, so it needs its own design:
timeouts, output limits, sanitization, and how it interacts with a
configuration file that someone copied from the internet.

### Configurable footer

The line under the card shows only the collection time in version 2.0 and
can only be turned on or off. The deferred version lets the configuration
set its content: free text and placeholders such as the collection time,
the version, the date or the value of an item.

Constraint that keeps it possible: `ui` receives the footer content as
data and does not compute it, so a configured footer replaces one input
and leaves the layout untouched.

### Prebuilt binaries and distribution packages

Release archives, an AUR package, a Nix package and similar. Version 2.0
is installed with `go install` only.

### Image logos

Logos drawn with the kitty graphics protocol or sixel. Version 2.0 has
text logos only.

### More distributions

Version 2.0 ships about twenty logos and falls back through `ID_LIKE` to
a generic one. Adding a distribution is a table entry and two logo files.

### Display refresh rate and scaling

Version 2.0 reads the resolution from sysfs. The refresh rate and the
scale factor need the display server and differ between X11 and Wayland
compositors.

## Non-goals

Deliberate decisions. Do not add them without an explicit request from the
maintainer.

- Live or continuously updating output. fetchify prints once and exits.
- Windows support.
- Escape sequences produced outside beautify, other than clearing the
  screen.
