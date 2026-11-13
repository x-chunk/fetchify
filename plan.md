# Implementation plan: fetchify 2.0

This plan describes the next version of fetchify: a card-style output built
on the `beautify` library, distribution logos, more facts, and a TOML
configuration file. Read `AGENTS.md` first; its rules apply to every step.
`goals.md` lists what is deliberately not part of this version.

## Contents

1. Scope and principles
2. Prerequisites
3. Target look
4. Global requirements
5. Repository layout
6. Milestone 0: scaffolding
7. Milestone 1: move to beautify without changing the look
8. Milestone 2: `sysinfo` rework
9. Milestone 3: default facts
10. Milestone 4: opt-in facts
11. Milestone 5: `distro`
12. Milestone 6: `config`
13. Milestone 7: `ui`
14. Milestone 8: command line
15. Milestone 9: documentation and release
16. Cross-cutting edge-case matrix
17. Final verification checklist
18. Decision log

---

## 1. Scope and principles

### 1.1 What is built

- A new default look: the logo of the running distribution next to a
  framed card that is split into titled sections, with usage bars, icons
  and an accent color taken from the distribution.
- About twenty distribution logos with a fallback chain.
- New facts: six shown by default, the rest opt-in.
- A TOML configuration file that controls sections, items and decoration.
- Parallel collection with per-item timeouts.

### 1.2 Principles that decide ambiguous cases

When this plan does not cover a situation, decide with these rules, in
this order:

1. **Never break the shell startup.** fetchify runs from shell rc files.
   A missing tool, file or sensor is never an error: the item is hidden
   or shown as unknown. Only invalid configuration and invalid flags stop
   the program.
2. **Fast by default.** A fact is on by default only when it is cheap,
   local and safe to show on a screenshot. Anything that touches the
   network, needs root or needs a vendor tool is opt-in.
3. **Nothing collected is wasted.** A fact that no configured item needs
   is not collected.
4. **Layers stay separate.** `sysinfo` returns data, `ui` decides how it
   looks, `main` writes. The rules of `AGENTS.md` keep applying.
5. **Presentation is built from beautify blocks.** If a block cannot do
   what the look needs, the gap is closed in beautify, not worked around
   with escape sequences in fetchify.

### 1.3 Out of scope

Other operating systems, named themes, items that run user commands,
image logos and prebuilt packages. See `goals.md`.

---

## 2. Prerequisites

This plan is executed after the beautify v1 plan is complete.

- beautify is implemented and tagged, including the additions recorded in
  its decision log for fetchify: `BorderSet.LeftJoin`/`RightJoin`, the
  `Sections` block, captions, label styles and footers on `Border` and
  `Sections`, `dra.Gradient`, `dra.RestyleFunc` and the gradient options.
- This plan was written against `plan.md` of beautify, not against its
  code. **Before starting each milestone, read the real API of beautify
  and adapt names and signatures.** Where the real API differs, follow
  the code and record the difference in section 18. Do not change the
  behavior described here to fit an API quirk without asking the
  maintainer.
- If something the look needs is missing from beautify, stop and add it
  to beautify first (principle 5).
- `go.xchunk.org` serves `go-import` meta tags for `go.xchunk.org/fetchify`
  and `go.xchunk.org/beautify`. This is outside both repositories; without
  it `go install` and the dependency on beautify do not resolve.

---

## 3. Target look

Default output on Ubuntu, 16-core desktop, wide terminal. Icons are shown
here as `*`; real output uses Nerd Font glyphs.

```
        .-/+oossssoo+/-.          ╭─ * System ─────────────────────────── alice@box ─╮
    `:+ssssssssssssssssss+:`      │ * OS        Ubuntu 24.04 LTS                     │
  -+ssssssssssssssssssyyssss+-    │ * Kernel    6.8.0-45-generic                     │
.ossssssssssssssssssdMMMNysssso.  │ * Uptime    3h 25m                               │
+ssssssssshmydMMMMMMMNddddyssss+  │ * Packages  1842 (dpkg), 12 (flatpak)            │
ossssssshNMMMyhhyyyyhmNMMMNhssso  │ * Shell     zsh                                  │
ossssssshNMMMyhhyyyyhmNMMMNhssso  │ * Terminal  kitty                                │
+ssssssssshmydMMMMMMMNddddyssss+  │ * DE        GNOME                                │
.ossssssssssssssssssdMMMNysssso.  │ * WM        Mutter  Wayland                      │
  -+ssssssssssssssssssyyssss+-    ├─ * Hardware ─────────────────────────────────────┤
    `:+ssssssssssssssssss+:`      │ * Host      ThinkPad X1 Carbon Gen 11            │
        .-/+oossssoo+/-.          │ * CPU       AMD Ryzen 7 5800X  16 cores          │
                                  │ * GPU       NVIDIA GeForce RTX 3070              │
                                  │ * Display   2560x1440                            │
                                  │ * Battery   ████████████░░░░   78%  charging     │
                                  ├─ * Storage ──────────────────────────────────────┤
                                  │ * RAM       ████████░░░░░░░░   8.0 /  32.0 GiB  25% │
                                  │ * Swap      ████░░░░░░░░░░░░   1.0 /   4.0 GiB  25% │
                                  │ * /         ███░░░░░░░░░░░░░   100 /   500 GiB  21%  ext4 │
                                  │ * /home     ████░░░░░░░░░░░░   250 /  1000 GiB  25%  btrfs │
                                  ╰───────────────────────────────── ● ● ● ● ● ● ● ● ─╯
                                    collected in 12 ms
```

The sketch shows the structure; the right edge is straight in real output.
Section 13 defines every element exactly.

Elements:

- **Logo** of the distribution, vertically centered against the card.
- **Card**: one frame, three sections by default. The first section title
  sits in the top edge, the others in dividers that connect to the frame.
- **Caption**: `user@host` at the right end of the top edge.
- **Rows**: icon, label, value. Labels are padded to one width across all
  sections so every value starts in the same column.
- **Usage rows**: a bar colored by fill level, then used, total and
  percent in aligned columns.
- **Badges** for states: session type, battery status, disabled swap.
- **Muted details** for secondary information: core count, file system.
- **Palette**: eight dots in the bottom edge.
- **Footer**: one muted line under the card.
- **Accent**: the frame, section titles, icons and labels use the accent
  color of the distribution.

---

## 4. Global requirements

- **Platform.** Linux only. Files that use Linux-only facilities carry a
  `//go:build linux` tag only when the package would otherwise not
  compile elsewhere; the program is not required to build on other
  systems.
- **Language and module.** The Go version beautify requires. The module
  path changes to `go.xchunk.org/fetchify/v2` in milestone 9, not before.
- **Dependencies.** `go.xchunk.org/beautify` and
  `github.com/pelletier/go-toml/v2`. Nothing else: `gopsutil` is removed
  in milestone 2. Tests use the standard library only.
- **License.** MIT. Third-party material keeps its notice (section 11.5).
- **Errors.** Collection never returns an error to the user. Configuration
  and flag errors go to stderr with exit code 2. A failed write is exit
  code 1.
- **Output.** `main` is the only code that writes, through one
  `beautify.Composer`. No package below `main` reads `os.Stdout`.
- **Documentation.** Doc comments per `AGENTS.md`. Every package has a
  `doc.go`.
- **Static checks.** `gofmt -l .` prints nothing. `go vet ./...` is clean.
- **Tests.** `go test -race ./...` passes. Table-driven tests are the
  default. Rendering tests use golden files (section 6.2). Collectors are
  tested against fixture trees, never against the machine that runs the
  tests, except for one smoke test per package.
- **Commits.** Atomic, per `AGENTS.md`. Each milestone lists its expected
  commits. Push to the working branch after each milestone.

---

## 5. Repository layout

```
go.mod
main.go                     wiring only: flags -> config -> collect -> ui -> Composer
internal/
  cli/        flag parsing, help and version text
  config/     Config type, defaults, TOML loading and validation
    default.toml            the documented default file, embedded
  sysinfo/    facts as data; parallel collection
    source.go               Source: file system, command runner, environment
    collect.go              Collect, timeouts
    fact_*.go               one file per fact or group of related facts
    testdata/               fixture trees per scenario
  distro/     os-release parsing, the distribution table, logos
    logos/                  <id>.txt, <id>_small.txt, NOTICE; embedded
  ui/         beautify Composables that make up the screen
    theme.go icons.go
    items.go                the item registry
    row.go bar.go badge.go  row building blocks
    card.go palette.go logo.go footer.go
    screen.go               responsive layout
    testdata/               golden files
  golden/     golden-file test helper
.github/workflows/ci.yml
```

`internal/render` is deleted in milestone 1. `install.sh` is deleted in
milestone 9.

Import rules:

- `sysinfo`, `distro`, `config` and `golden` import nothing from this
  module, with one exception: `config` may import `distro` and `sysinfo`
  for their identifier types.
- `ui` imports `sysinfo`, `distro`, `config` and beautify.
- `cli` imports `config`.
- `main` imports everything.
- Only `ui` and `main` import beautify.

---

## 6. Milestone 0: scaffolding

### 6.1 CI

`.github/workflows/ci.yml`: on push and pull request, on `ubuntu-latest`,
run the `gofmt` check, `go vet ./...` and `go test -race ./...`.

### 6.2 Golden helper

`internal/golden` with one function, the same contract as the helper in
beautify (which is internal there and cannot be imported):

```go
// Assert compares got with testdata/<name>.golden relative to the test's
// package directory. With the -update flag it writes got instead.
func Assert(t *testing.T, name string, got []byte)
```

- Golden files store raw bytes, including escape sequences.
- A mismatch prints both values with `%q`.
- A missing file fails with a message that says to run with `-update`.
- Review golden diffs by hand before committing; never regenerate blindly.

### 6.3 Commits

1. `Add CI workflow`
2. `Add golden-file test helper`

---

## 7. Milestone 1: move to beautify without changing the look

A mechanical step that proves the dependency works before anything else
changes. Output after this milestone equals the output before it, byte
for byte, for the fixtures in `internal/ui`.

### 7.1 Work

- Add the dependency on beautify.
- Rewrite `internal/ui` against beautify using the table in the
  "Migration to beautify" section of `AGENTS.md`: spans become models
  from `ctx.Factory()`, `Build` becomes a function that returns a
  `beautify.Template`, the logo and the info column become a `blocks.Row`.
- Replace the write in `main.run` with a `beautify.Composer` that has a
  fixed `ProfileANSI16`, matching current behavior. Profile detection
  comes in milestone 8.
- Delete `internal/render` and the migration section of `AGENTS.md`;
  update the architecture section in the same commit.

### 7.2 Checks

- Before deleting `internal/render`, capture the output of the existing
  `ui` tests as golden files. After the rewrite the same fixtures produce
  the same bytes.

### 7.3 Commits

1. `Pin the current screen with golden files`
2. `Add beautify dependency`
3. `Rebuild ui on beautify blocks`
4. `Render through beautify.Composer`
5. `Remove internal/render`

---

## 8. Milestone 2: `sysinfo` rework

### 8.1 Facts

A fact is one independently collectable piece of data.

```go
// Fact identifies a piece of data sysinfo can collect.
type Fact string

// State says whether a fact was found.
type State uint8

const (
    Unknown State = iota // not requested, failed or timed out
    Present              // collected
    Absent               // does not exist on this machine
)

// Value is a fact of type T with its state. The zero value is Unknown.
type Value[T any] struct {
    State State
    V     T
}
```

`Absent` and `Unknown` are different on purpose and the presentation
depends on it: a desktop without a battery is `Absent` and the row is
hidden; a battery whose files cannot be read is `Unknown` and the row
says so.

`Info` keeps one typed field per fact, each a `Value[T]`. Field types
carry raw data: sizes in bytes, durations as `time.Duration`, lists as
slices. No field holds a formatted string that the UI would have to
parse.

### 8.2 Source

Collectors never touch the machine directly. They read through a
`Source`:

```go
// Source is everything a collector may read. Tests replace it.
type Source struct {
    FS     fs.FS                                  // rooted at "/"
    Run    func(ctx context.Context, name string, args ...string) ([]byte, error)
    Getenv func(string) string
    Statfs func(path string) (FSStat, error)
    Now    func() time.Time
    // identity of the process: effective UID, PID and user lookup
}

func SystemSource() Source
```

- `FS` is `os.DirFS("/")` in production. Paths in collectors are written
  without the leading slash.
- `Run` uses `exec.CommandContext`, so a timeout kills the command.
  Standard error is discarded. Commands are run with their plain name and
  resolved through `PATH`; never through a shell.
- A collector that needs something a `Source` cannot give gets a new
  field on `Source`, not a direct call.

### 8.3 Collection

```go
// Options controls a collection run.
type Options struct {
    Timeout  time.Duration            // default per-fact limit
    Timeouts map[Fact]time.Duration   // per-fact overrides
    // collector settings that come from configuration, for example the
    // list of ignored mount points
}

// Collect gathers the given facts concurrently. It never fails.
func Collect(ctx context.Context, src Source, facts []Fact, opts Options) Info

// Facts returns every fact sysinfo knows, sorted.
func Facts() []Fact
```

- Each requested fact runs in its own goroutine with its own deadline.
- A collector that returns after its deadline is ignored and the fact is
  `Unknown`. `Collect` does not wait for it: it returns when every fact is
  done or expired. Collectors must honor `ctx` so goroutines do not leak
  past the commands they started.
- A panic in a collector is recovered and the fact is `Unknown`. This is
  the one place where recovery is right: a bug in one parser must not
  take down the shell startup.
- Facts not requested are not collected and stay `Unknown`.
- Duplicate facts in the request are collected once. Unknown identifiers
  are ignored; validation is the job of `config`.
- Each collector writes only its own field, so no locking is needed
  beyond waiting for completion. Collectors that share an expensive read
  (for example the process table) share it through a `sync.Once` inside
  the run.
- Default `Timeout`: 200 ms.

Performance target: with the default configuration, on a machine with a
warm file cache, `Collect` finishes in under 30 ms. This is a goal for
benchmarks, not an assertion in tests.

### 8.4 Removing gopsutil

Every current use has a direct Linux source (section 9). Remove the
dependency once the last collector is ported.

### 8.5 Tests

- `Collect` with a collector that blocks forever returns within the
  timeout plus a small margin, and the fact is `Unknown`.
- `Collect` with a panicking collector returns and the other facts are
  intact.
- Only requested facts are collected (count calls with fake collectors).
- Race detector: collect all facts from a fixture source many times.
- Per-fact tests: a fixture tree per scenario under `testdata/`, loaded
  with `os.DirFS` or `fstest.MapFS`, asserting the typed value and state.

### 8.6 Commits

1. `Add Fact, State and Value to sysinfo`
2. `Add Source to sysinfo`
3. `Collect facts concurrently with timeouts`
4. One commit per ported collector (section 9)
5. `Remove gopsutil dependency`

---

## 9. Milestone 3: default facts

For every fact: the source, the data type, and when it is `Absent`.
"Fixtures" lists the scenarios that need a test tree.

### 9.1 System

| Fact | Source | Data | Notes |
|------|--------|------|-------|
| `user` | process identity | name, UID | name falls back to `$USER`, then to the UID as text |
| `hostname` | `proc/sys/kernel/hostname` | string | |
| `os` | `etc/os-release`, then `usr/lib/os-release` | `distro.Release` (section 11.1) | `Unknown` when neither file exists |
| `kernel` | `proc/sys/kernel/osrelease` | string | |
| `uptime` | first field of `proc/uptime` | `time.Duration` | truncated to seconds |
| `packages` | see 9.2 | `[]PackageCount{Manager, Count}` | `Absent` when no manager is found |
| `shell` | `$SHELL` | base name | `Unknown` when unset |
| `terminal` | see 9.3 | string | `Absent` on a Linux console or without a terminal |
| `session` | `$XDG_SESSION_TYPE`, `$WAYLAND_DISPLAY`, `$DISPLAY` | `X11`, `Wayland` or none | `Absent` without a graphical session |
| `de` | `$XDG_CURRENT_DESKTOP`, then `$DESKTOP_SESSION` | string | `Absent` when `session` is absent; first entry of a colon-separated list |
| `wm` | see 9.4 | string | `Absent` when `session` is absent |

### 9.2 Packages

Count per manager. A manager is included when its database exists. Prefer
reading the database over running the tool: it is faster and needs no
timeout budget.

| Manager | How to count |
|---------|--------------|
| `pacman` | directories in `var/lib/pacman/local` |
| `dpkg` | paragraphs of `var/lib/dpkg/status` whose `Status:` line ends with `installed` |
| `rpm` | `rpm -qa` line count; included when `var/lib/rpm` or `usr/lib/sysimage/rpm` exists |
| `xbps` | `xbps-query -l` line count; included when `var/db/xbps` exists |
| `apk` | lines starting with `P:` in `lib/apk/db/installed` |
| `portage` | package directories two levels under `var/db/pkg` |
| `nix` | `nix-store -q --requisites /run/current-system/sw` line count; included when `nix/store` exists |
| `flatpak` | directories in `var/lib/flatpak/app` plus `<home>/.local/share/flatpak/app` |
| `snap` | directories in `snap` except `bin` and `README` |

- Managers are reported in the order of this table. Managers with a count
  of 0 are dropped.
- A manager whose tool fails or times out is dropped, not reported as 0.

### 9.3 Terminal

1. If `$TERM_PROGRAM` is set, use it (strip a `.app` suffix).
2. Otherwise check the variables that terminals set for themselves:
   `KITTY_WINDOW_ID` (kitty), `ALACRITTY_WINDOW_ID` (Alacritty),
   `WEZTERM_EXECUTABLE` (WezTerm), `KONSOLE_VERSION` (Konsole),
   `GNOME_TERMINAL_SCREEN` (GNOME Terminal), `FOOT_*`/`TERM=foot` (foot),
   `TILIX_ID` (Tilix), `GHOSTTY_RESOURCES_DIR` (Ghostty).
3. Otherwise walk up the process tree from the parent through
   `proc/<pid>/stat` and `proc/<pid>/comm`, skipping shells and wrappers
   (`sh`, `bash`, `zsh`, `fish`, `dash`, `nu`, `sudo`, `su`, `login`),
   and take the first other process. Stop at PID 1 or after 16 steps.
4. Multiplexers (`tmux`, `screen`, `zellij`) are reported as found: the
   real terminal is not reachable through the process tree.
5. Over SSH (`$SSH_TTY` set) the walk ends at `sshd`: report `Absent`.

### 9.4 Window manager

- Wayland: scan process names for known compositors and map them to
  display names: `sway`, `Hyprland`, `kwin_wayland` (KWin), `gnome-shell`
  (Mutter), `river`, `niri`, `labwc`, `wayfire`, `weston`, `cosmic-comp`
  (COSMIC). First match in this order.
- X11: `xprop -root _NET_SUPPORTING_WM_CHECK`, then `_NET_WM_NAME` of that
  window; if `xprop` is missing, `wmctrl -m`; if both are missing, scan
  process names for known window managers (`i3`, `bspwm`, `openbox`,
  `awesome`, `xfwm4`, `kwin_x11`, `dwm`, `qtile`, `xmonad`, `herbstluftwm`).
- Nothing found: `Unknown`.

### 9.5 Hardware

| Fact | Source | Data | Notes |
|------|--------|------|-------|
| `host` | `sys/devices/virtual/dmi/id/product_name` and `product_version`; else `sys/firmware/devicetree/base/model` | string | see below |
| `cpu` | `proc/cpuinfo` | model, logical count | see below |
| `gpu` | see 9.6 | `[]string` | `Absent` when there are no display devices |
| `display` | `sys/class/drm/card*-*/status` and `modes` | `[]Display{Width, Height}` | connected outputs only; first mode; `Absent` when none is connected |
| `battery` | `sys/class/power_supply/*/` where `type` is `Battery` and `scope` is not `Device` | `[]Battery{Percent, Status}` | `Absent` when there is none |

- `host`: discard placeholder values, compared case-insensitively:
  `To Be Filled By O.E.M.`, `System Product Name`, `Default string`,
  `Not Applicable`, `None`, `Unknown`, empty. Append the version only when
  it is not a placeholder and not already contained in the name (Lenovo
  stores the readable name in `product_version`: when the name looks like
  a model code and the version is readable, use the version). When
  everything is a placeholder: `Unknown`.
- `cpu`: model from the first `model name` line; on ARM, where it is
  missing, from `Hardware` or `Model`. Logical count is the number of
  `processor` entries. Clean the model for display in `ui`, not here.
- `battery.Status`: the content of `status`, one of `Charging`,
  `Discharging`, `Full`, `Not charging`, `Unknown`.

### 9.6 GPU

1. For each `sys/class/drm/card<N>` (no connector suffix), read
   `device/vendor`, `device/device` and `device/class`. Keep classes
   `0x0300` (VGA), `0x0302` (3D) and `0x0380` (display). The current
   `lspci | grep VGA` misses 3D controllers, which is how discrete GPUs
   appear on most laptops.
2. Resolve names from the first existing database:
   `usr/share/hwdata/pci.ids`, `usr/share/misc/pci.ids`,
   `usr/share/pci.ids`. Format: `<vendor name> <device name>`.
3. If no database exists, run `lspci -mm` and take the matching devices.
4. If that fails too, report the vendor by its well-known ID (`10de`
   NVIDIA, `1002` AMD, `8086` Intel) followed by the device ID.
5. Deduplicate by PCI address. Keep bus order.

### 9.7 Storage

| Fact | Source | Data | Notes |
|------|--------|------|-------|
| `memory` | `proc/meminfo` | `Usage{Used, Total}` | used is `MemTotal - MemAvailable` |
| `swap` | `proc/meminfo` | `Usage{Used, Total}` | `Absent` when `SwapTotal` is 0 |
| `disks` | `proc/self/mounts` and `Statfs` | `[]Disk{Mountpoint, Filesystem, Usage}` | see below |

- `Usage` drops the stored percent: the UI computes it from `Used` and
  `Total`. For disks, `Used` is `total - free` and the percent uses
  `used / (used + available)` so it matches `df`; keep `Available` in the
  type for that.
- Disks, in mount order:
  - Skip pseudo and image file systems by type: `proc`, `sysfs`, `tmpfs`,
    `devtmpfs`, `devpts`, `cgroup`, `cgroup2`, `overlay`, `squashfs`,
    `efivarfs`, `fuse.portal`, `autofs`, `ramfs`, `debugfs`, `tracefs`,
    `securityfs`, `pstore`, `bpf`, `configfs`, `mqueue`, `hugetlbfs`,
    `binfmt_misc`, `nsfs`, `fusectl`, `rpc_pipefs`.
  - Skip mount points inside the ignore list (matched by path component).
    Default list: `/boot`, `/efi`, `/run`, `/var`, `/tmp`, `/dev`, `/sys`,
    `/proc`, `/snap`, `/nix/store`. `/mnt` is no longer ignored: external
    drives are worth showing.
  - Deduplicate by device: keep the mount with the shortest mount point.
    This collapses btrfs subvolumes and bind mounts.
  - Skip file systems with a total size of 0 and ones `Statfs` fails on.
  - Decode the octal escapes of `proc/self/mounts` (`\040` is a space).
- `Absent` when the list is empty.

### 9.8 Fixtures

At least: a typical Ubuntu desktop, an Arch laptop with a battery and two
GPUs, a headless server over SSH, a container (no DMI, no DRM, overlay
root), a Raspberry Pi (device tree, no DMI), a machine with btrfs
subvolumes, a machine with every placeholder DMI value.

### 9.9 Commits

One commit per fact or tightly related group, each with its fixtures and
tests, in the order of this section.

---

## 10. Milestone 4: opt-in facts

Off by default. Each becomes available as an item of the same name
(section 13.2). Collected only when configured.

| Fact | Source | Data | Why opt-in |
|------|--------|------|------------|
| `local_ip` | `net.InterfaceAddrs`, first global unicast address of an interface that is up and not loopback | address, interface | visible on screenshots |
| `public_ip` | HTTPS request to a configurable endpoint (`items.public_ip.url`, default `https://api.ipify.org`) | address | network; default timeout 1 s |
| `network` | default-route interface from `proc/net/route`; SSID through `iw dev <if> link` or `nmcli` | interface, type, SSID | privacy |
| `bluetooth` | `sys/class/bluetooth` and `rfkill` state | on or off | niche |
| `cpu_temp` | `sys/class/hwmon/*/`: `k10temp`, `coretemp`, `zenpower`, then `sys/class/thermal/thermal_zone*/` with type `x86_pkg_temp` or `cpu-thermal` | degrees Celsius | sensors vary |
| `gpu_temp` | `hwmon` under the GPU device; `nvidia-smi` for NVIDIA | degrees Celsius | vendor tools |
| `cpu_usage` | two samples of `proc/stat` 100 ms apart | percent | costs 100 ms |
| `load` | `proc/loadavg` | three numbers | niche |
| `gpu_driver` | `sys/class/drm/card<N>/device/driver` link target | string per GPU | niche |
| `gpu_memory` | `mem_info_vram_total`/`_used` in sysfs (AMD); `nvidia-smi` (NVIDIA) | `Usage` per GPU | vendor tools |
| `memory_type` | `dmidecode -t memory` | `[]string` | needs root; `Absent` when not root |
| `board` | `sys/devices/virtual/dmi/id/board_vendor` and `board_name` | string | niche |
| `bios` | `bios_vendor`, `bios_version`, `bios_date` | string | niche |
| `init` | `proc/1/comm` | string | niche |
| `processes` | numeric directories in `proc` | count | niche |
| `users` | distinct users with a session in `run/utmp`, or `who` | count | niche |
| `locale` | `$LC_ALL`, `$LC_MESSAGES`, `$LANG` | string | niche |
| `timezone` | target of `etc/localtime` | string | niche |
| `date` | `Source.Now` | time | niche |
| `theme`, `icon_theme`, `font`, `cursor` | GTK settings: `gsettings` for GNOME-family, `kreadconfig` for KDE, then `settings.ini` of GTK 3/4 | string | long, desktop-specific |
| `terminal_font` | the configuration of the detected terminal, for the terminals where that is one known file (kitty, Alacritty, foot, WezTerm, Ghostty) | string | fragile |
| `media` | `playerctl metadata` | artist, title, status | needs a tool |

Rules:

- Every opt-in collector follows 8.2 and 8.3: no direct access, honors
  the context, never fails.
- A fact whose tool is not installed is `Unknown`. A fact that does not
  apply (no Bluetooth adapter, no player running, not root) is `Absent`.
- `public_ip` sends no data besides the request itself, sets a generic
  `User-Agent` of `fetchify`, and accepts a response only when it parses
  as an IP address.
- `cpu_usage` is the only collector allowed to sleep.

Commits: one per fact or related group, with fixtures and tests.

---

## 11. Milestone 5: `distro`

### 11.1 Release

```go
// Release is the parsed content of os-release.
type Release struct {
    ID         string   // "ubuntu"
    IDLike     []string // "debian"
    Name       string   // "Ubuntu"
    PrettyName string   // "Ubuntu 24.04 LTS"
    VersionID  string   // "24.04"
    BuildID    string
    VariantID  string
}

func ParseRelease(data []byte) Release
```

- Format per `os-release(5)`: `KEY=value`, values optionally in single or
  double quotes, backslash escapes inside double quotes, `#` comments.
- Unknown keys, malformed lines and duplicate keys (last wins) are
  tolerated. Parsing never fails.
- `ID` and `IDLike` are lower-cased.

### 11.2 Distribution table

```go
// Distro is what fetchify knows about a distribution.
type Distro struct {
    ID      string       // lookup key, lower case
    Name    string       // display name, used when os-release has none
    Colors  []dra.Color  // logo palette, in marker order
    Accent  dra.Color    // color of the frame, titles, icons and labels
    Icon    rune         // Nerd Font glyph of the distribution
}

func Lookup(r Release) Distro
func ByID(id string) (Distro, bool)
func IDs() []string
func (d Distro) Logo(size Size) Logo   // Full or Small
```

`distro` importing `dra` for colors is the one exception to "only `ui`
and `main` import beautify": colors are data here. If that turns out to
be awkward, store colors as hex strings and convert in `ui`; record the
choice in section 18.

First version, 22 distributions plus two generic entries:

`arch`, `debian`, `ubuntu`, `fedora`, `linuxmint`, `manjaro`, `opensuse`,
`nixos`, `gentoo`, `void`, `alpine`, `endeavouros`, `cachyos`, `pop`,
`kali`, `rocky`, `almalinux`, `rhel`, `centos`, `zorin`, `elementary`,
`artix`; `linux` (Tux, the generic fallback) and `koala` (the fetchify
mascot, never chosen automatically).

Aliases resolve several identifiers to one entry: `opensuse-tumbleweed`,
`opensuse-leap`, `opensuse-slowroll` and `sles` to `opensuse`; `archarm`
to `arch`; `raspbian` to `debian`; `ol` to `rhel`.

### 11.3 Lookup

`Lookup` returns the first match of:

1. `ID`, directly or through an alias.
2. Each entry of `IDLike`, in order, directly or through an alias.
3. `linux`.

So an unknown Ubuntu derivative gets the Ubuntu logo and an unknown
distribution gets Tux. `Lookup` never fails.

### 11.4 Logo format

One text file per logo, embedded with `embed`.

- Color markers `${1}` to `${9}` switch to the palette color with that
  number for the text that follows, until the next marker. Text before
  the first marker uses color 1. `$$` is a literal dollar sign.
- A marker that refers to a color the distribution does not have falls
  back to color 1.
- Lines are not padded in the file; the loader pads nothing either. The
  logo is a left-aligned block.
- No tabs, no trailing spaces, no control characters. A test enforces it.
- Full logos are at most 40 columns by 20 lines. Small logos are at most
  16 columns by 8 lines. A test enforces it.

```go
// Logo is a parsed logo: lines of runs that reference palette colors.
type Logo struct { /* lines of {text, color index} */ }

func (l Logo) Width() int
func (l Logo) Height() int
```

`ui` turns a `Logo` into a `dra.Model` with a palette and a colormap.

### 11.5 Sources and attribution

- Take existing ASCII logos from neofetch and fastfetch (both MIT) and
  convert them to the format above. Draw the small variants that do not
  exist there.
- `internal/distro/logos/NOTICE` lists, per logo, where it came from and
  reproduces the MIT notices of both projects.
- `README.md` states that distribution names and logos are trademarks of
  their owners and are used only to identify the running system.
- The koala is fetchify's own and is redrawn without the title and
  version lines it carries today.

### 11.6 Tests

- `ParseRelease`: real `os-release` files of every listed distribution
  as fixtures, plus quoting, escapes, comments, empty file, binary junk.
- `Lookup`: direct hit, alias, `ID_LIKE` chain, unknown, empty release.
- Table integrity: every entry has a full and a small logo, at least one
  color, an accent, a unique ID; every alias points to an existing entry;
  every marker in every logo refers to an existing color.
- Format rules of 11.4 for every logo.

### 11.7 Commits

1. `Add os-release parsing to distro`
2. `Add the logo format and loader`
3. `Add the distribution table and lookup`
4. Logos in groups of related distributions, each with its NOTICE entries
5. `Read the OS fact through distro.Release`

---

## 12. Milestone 6: `config`

### 12.1 Location and precedence

- Default path: `$XDG_CONFIG_HOME/fetchify/config.toml`, or
  `~/.config/fetchify/config.toml` when the variable is unset or not an
  absolute path.
- `--config <path>` reads that file instead.
- The default file not existing is normal: defaults apply. A path given
  with `--config` that cannot be read is an error.
- Precedence: flags, then the file, then defaults.

### 12.2 Schema

`default.toml`, embedded, is the documented default configuration. It is
what `--print-config` prints and it must decode to exactly `Defaults()`;
a test enforces this, so the file can never drift from the code.

```toml
clear      = false     # clear the screen before printing
footer     = true      # the line under the card
timeout_ms = 200       # default time limit per item

[logo]
name      = "auto"     # auto | a distribution id | "koala" | "none"
size      = "auto"     # auto | full | small
gradient  = false      # recolor the logo with a gradient
colors    = []         # gradient stops; empty = the logo's own colors
direction = "vertical" # vertical | horizontal | diagonal

[colors]
mode   = "auto"        # auto | truecolor | 256 | 16 | none
accent = "auto"        # auto = from the distribution, or "#rrggbb"

[border]
style     = "rounded"  # rounded | square | double | heavy | ascii
color     = "accent"   # accent | logo | rainbow | custom
colors    = []         # stops for color = "custom"
direction = "vertical" # vertical | horizontal | diagonal

[icons]
enabled = true         # needs a Nerd Font

[bars]
style      = "block"   # block | thin
width      = 16
thresholds = [60, 85]  # percent where the color turns to warning and danger

[palette]
style = "dots"         # dots | blocks | none

[[group]]
title = "System"
icon  = "auto"
items = ["os", "kernel", "uptime", "packages", "shell", "terminal", "de", "wm"]

[[group]]
title = "Hardware"
icon  = "auto"
items = ["host", "cpu", "gpu", "display", "battery"]

[[group]]
title = "Storage"
icon  = "auto"
items = ["memory", "swap", "disks"]

[items.disks]
ignore = ["/boot", "/efi", "/run", "/var", "/tmp", "/dev", "/sys", "/proc", "/snap", "/nix/store"]
```

Per-item tables `[items.<id>]`, valid for every item:

| Key | Meaning |
|-----|---------|
| `label` | replaces the label |
| `icon` | replaces the icon; `""` removes it |
| `timeout_ms` | replaces the time limit for this item |

Item-specific keys: `items.disks.ignore`, `items.memory.show_type`
(adds the `memory_type` fact), `items.public_ip.url`.

### 12.3 Semantics

- **Groups replace, they do not merge.** If the file defines any
  `[[group]]`, the three default groups are not added.
- **Everything else merges.** A key that is absent keeps its default.
  An array that is present replaces the default array as a whole.
- An item appears where the file puts it and is collected only if it is
  in a group.
- `group.icon = "auto"` takes the icon registered for the default group
  of the same title, and no icon for other titles.

### 12.4 Validation

Loading stops at invalid input. All problems found are reported, one per
line, then the program exits with code 2 and prints nothing else.

Format: `fetchify: <path>:<line>:<column>: <message>`. The position is
included whenever the TOML library provides it; a message without a
position still names the full key, for example `border.color`.

| Problem | Message names |
|---------|---------------|
| TOML syntax error | what the parser reports |
| unknown key or table | the key |
| wrong type | the key, the expected type |
| value outside its enumeration | the key, the value, the allowed values |
| unknown item identifier | the identifier, and the closest known one when the edit distance is at most 2 |
| item listed twice, in one group or in two | the identifier and both groups |
| `[items.<id>]` for an unknown item | the identifier |
| item-specific key on another item | the key and the item |
| color not of the form `#rgb` or `#rrggbb` | the key and the value |
| `colors` empty where stops are required (`border.color = "custom"`) | the key |
| `bars.width` outside 4-40 | the value |
| `bars.thresholds` not two ascending numbers within 0-100 | the value |
| `timeout_ms` outside 1-10000 | the value |
| `logo.name` not `auto`, `none` or a known logo | the value and how to list logos |
| group with an empty title | the group index |

A group with no items is valid and is not drawn. A file with
`group = []` is valid and draws the logo only.

Use the strict mode of the TOML library for unknown keys so positions are
available. Known item and logo identifiers are passed in by the caller so
`config` does not import `ui`.

### 12.5 API

```go
func Defaults() Config
func DefaultPath(getenv func(string) string, home string) string
func Load(path string, known Known) (Config, error)   // error lists every problem
func DefaultTOML() []byte
```

The error type exposes the individual problems so `main` can print them
one per line and tests can assert on them.

### 12.6 Tests

- `default.toml` decodes to `Defaults()`.
- Every row of the validation table, including the position.
- Several problems in one file are all reported.
- Merge rules: a file with one key; a file with one group; an empty file.
- Path resolution with and without `XDG_CONFIG_HOME`, with a relative
  value, with no home directory.
- Fuzz test `FuzzLoad`: never panics.

### 12.7 Commits

1. `Add Config and defaults`
2. `Add the documented default configuration file`
3. `Load configuration from TOML`
4. `Validate configuration and report every problem`

---

## 13. Milestone 7: `ui`

Everything in this milestone is built from beautify blocks and `dra`
operations. Names below follow the beautify plan; adapt them to the real
API (section 2).

### 13.1 Theme

One struct resolved once per run from the configuration and the
distribution:

| Role | Default |
|------|---------|
| `Accent` | `colors.accent`, or the accent of the distribution |
| `Frame` | the accent, or a gradient (13.9) |
| `Title` | accent, bold |
| `Label` | accent |
| `Icon` | accent |
| `Value` | zero style: the terminal's own foreground |
| `Muted` | bright black (palette index 8) |
| `Caption` | bold; red and bold when the process runs as root |
| `Good`, `Warning`, `Danger` | green, yellow, red (palette indexes 2, 3, 1) |

`Good`, `Warning`, `Danger`, `Muted` and the palette dots use the 16-color
palette on purpose, so they follow the user's terminal theme. Only the
accent, the logo and gradients are 24-bit; beautify downgrades them for
weaker terminals.

No section of the UI uses a color directly; everything goes through the
theme.

### 13.2 Items

An item is one configurable entry of a group. The registry maps each
identifier to its definition:

```go
type item struct {
    id     string
    label  string           // default label
    icon   rune             // default icon
    facts  []sysinfo.Fact   // what it needs collected
    rows   func(info sysinfo.Info, env renderEnv) []row
}
```

- `rows` returns zero rows when every needed fact is `Absent`: the item
  disappears. It returns one row with a muted `unknown` value when the
  fact is `Unknown`.
- An item can produce several rows (`gpu`, `disks`, `battery`).
- A group whose items produce no rows is not drawn, together with its
  title.
- `ui` exports the item identifiers and, for a configuration, the set of
  facts to collect, so `main` can call `sysinfo.Collect` with exactly
  that set.

Default items:

| Item | Label | Value |
|------|-------|-------|
| `os` | OS | `PrettyName`, else `Name` and `VersionID`. On Arch a muted `btw` follows. The icon is the icon of the distribution. |
| `kernel` | Kernel | the release |
| `uptime` | Uptime | `2d 3h 25m`; the two largest non-zero units below a day are kept (`3h 25m`), under a minute `42s` |
| `packages` | Packages | `1842 (dpkg), 12 (flatpak)`; the manager names muted |
| `shell` | Shell | name |
| `terminal` | Terminal | name |
| `de` | DE | name; hidden when equal to the window manager, ignoring case |
| `wm` | WM | name, then a badge with the session type |
| `host` | Host | model |
| `cpu` | CPU | cleaned model, then muted `16 cores` |
| `gpu` | GPU | one row per device, cleaned name |
| `display` | Display | `2560x1440`; several joined with `, ` |
| `battery` | Battery | usage row without sizes: bar, percent, status badge; one row per battery |
| `memory` | RAM | usage row; with `show_type`, the types as a muted detail |
| `swap` | Swap | usage row; a muted `disabled` badge when `Absent` |
| `disks` | the mount point | usage row with the file system as a muted detail; one row per disk |

Name cleaning, applied in `ui`:

- CPU: remove `(R)`, `(TM)`, `(tm)`, `CPU`, `Processor`, a trailing
  `@ <n>GHz`, a trailing `<n>-Core Processor`, and collapse spaces.
- GPU: remove `Corporation`, `Inc.`, `Co., Ltd.`, a trailing `(rev xx)`;
  when the name has a bracketed part (`[GeForce RTX 3070]`), keep the
  vendor and the bracket content.

Opt-in items use the same shape. Their labels and formats: temperatures
as `54°C` colored by `Good`/`Warning`/`Danger` at 70 and 85; `cpu_usage`
and `gpu_memory` as usage rows; everything else as plain text. Each gets
a default group for documentation only; it appears wherever the
configuration puts it.

### 13.3 Rows

A row is `icon  label  value`.

- The label column is as wide as the widest label of all drawn rows in
  all sections, so values align across the whole card.
- With icons enabled the icon takes one cell followed by one space. A row
  without an icon keeps the two cells empty so labels stay aligned.
- With icons disabled the icon column does not exist.
- Two spaces separate the label column from the value.
- A value is one line. Newlines cannot occur: beautify sanitizes text.

### 13.4 Usage rows and bars

A usage row is `bar  used / total unit  percent  detail`.

- Within a section the columns of all usage rows align: bars have equal
  width; `used`, `total` and the percent are right-aligned in columns as
  wide as the widest entry.
- **Unit.** One unit per row, chosen by the total: `MiB` below 1 GiB,
  `TiB` from 1024 GiB, otherwise `GiB`. All units are binary.
- **Numbers.** One decimal below 100, none from 100 (`8.0`, `250`).
  Percent is rounded to the nearest integer and clamped to 0-100.
- **Bar.** `bars.width` cells. Filled cells are
  `round(width * percent / 100)`, at least 1 when the percent is above 0
  and at most `width - 1` when it is below 100, so "almost empty" and
  "almost full" are visible.

| `bars.style` | Filled | Empty |
|--------------|--------|-------|
| `block` | `█` in the level color | `░` muted |
| `thin` | `━` in the level color | `━` muted |

- **Level color.** Below `thresholds[0]`: `Good`. Below `thresholds[1]`:
  `Warning`. Otherwise `Danger`. The whole filled part has one color. The
  percent text is not colored.
- For the battery the scale is inverted: a low charge is the danger. Use
  thresholds 20 and 40 from the low end, fixed.
- The bar is a custom Composable built with `dra.Repeat` and `dra.Concat`.
  It honors the width contract of beautify: with less room than its
  width, it shrinks.

### 13.5 Badges

`blocks.Badge` with one of three looks: neutral (muted, reversed) for the
session type, `Good` for `charging` and `full`, `Warning` for
`discharging` below 20 percent, neutral otherwise, and neutral for
`disabled`. Badge text is lower case except protocol names (`X11`,
`Wayland`).

### 13.6 Card

`blocks.Sections` with one section per drawn group.

- Section title: the group icon (when icons are enabled and the group has
  one), a space, the group title, in the `Title` style.
- Caption: `user@host` in the `Caption` style. When either part is
  unknown, the known part alone; when both are, no caption.
- Border set from `border.style`. Padding 1, 0.
- Footer of the frame: the palette, when `palette.style` is `dots`.
- No groups drawn: no card at all.

### 13.7 Palette

- `dots`: eight `●` separated by single spaces, in palette colors 0-7,
  as the footer model of the card.
- `blocks`: two identical rows of eight `███`, in palette colors 0-7,
  under the card, indented by one cell. This is the old look.
- `none`: nothing.

Palette colors are indexed, never 24-bit: the palette exists to show the
terminal's own colors.

### 13.8 Logo

- `logo.name = "auto"`: `distro.Lookup` of the collected release. A
  distribution identifier or `koala`: that logo. `none`: no logo.
- The `distro.Logo` becomes one `dra.Model`: palette from the colors of
  the distribution, colormap from the runs.
- `logo.gradient = true` recolors the whole logo with
  `dra.RestyleFunc` and a `dra.Gradient` whose stops are `logo.colors`,
  or the colors of the distribution when that is empty, along
  `logo.direction`. Spaces are not recolored (no background is set, so
  this is invisible either way).

### 13.9 Frame color

| `border.color` | Frame |
|----------------|-------|
| `accent` | solid accent |
| `logo` | gradient through the colors of the distribution; solid when it has one color |
| `rainbow` | gradient through `#ff5f5f`, `#ffbf5f`, `#ffff5f`, `#5fff5f`, `#5fffff`, `#5f5fff`, `#ff5fff` |
| `custom` | gradient through `border.colors`; solid when there is one |

Gradients use `SectionsBorderGradient` along `border.direction`. Section
titles keep the `Title` style; they do not follow the gradient.

### 13.10 Footer

`footer = true`: one line under the card, indented by two cells, in the
`Muted` style: `collected in <n> ms`, where the time is that of
`sysinfo.Collect`, rounded to milliseconds. The version is not shown
here; it is available through `--version`.

The time is passed into `ui` as data. `ui` does not measure anything.

### 13.11 Screen and narrow terminals

The screen is a row: the logo, a gap of two cells, and a column of the
card, the block palette (if configured) and the footer. The logo is
centered vertically against the column; when the logo is taller, the
column is centered against the logo.

The available width comes from the beautify context. When it is
unbounded (output is not a terminal and `COLUMNS` is unset), use the
natural layout. Otherwise pick the first layout that fits, in this
order:

1. Full logo and natural card.
2. Small logo and natural card.
3. No logo and natural card.
4. No logo, and the card shrinks in these steps, stopping as soon as it
   fits:
   1. Muted details are dropped (core count, file system, manager names).
   2. Bars shrink, down to 6 cells.
   3. Bars are dropped.
   4. Values are truncated with an ellipsis.
5. If it still does not fit, beautify truncates the result to the bound.

`logo.size = "full"` or `"small"` pins the logo size: steps 1 and 2 are
replaced by that one size, and step 3 still removes the logo when it does
not fit. `logo.name = "none"` starts at step 3.

Implement the screen as one Composable that measures candidates by
rendering them. Rendering is cheap and pure; do not predict widths by
hand.

### 13.12 Icons

- One table maps item identifiers, group titles and distributions to Nerd
  Font glyphs. Choose glyphs from the stable ranges of Nerd Fonts v3
  (Font Awesome, Material Design, Font Logos, Octicons) and note the
  glyph name next to each code point in a comment.
- Every icon is treated as one cell wide.
- `icons.enabled = false` or `--no-icons` removes the icon column, the
  group icons and the distribution icon. Nothing else changes.
- Nerd Font support cannot be detected. Icons are on by default by
  decision of the maintainer; `README.md` explains how to turn them off.

### 13.13 Tests

- Golden tests of the whole screen at `ProfilePlain` and
  `ProfileTrueColor` for: each fixture machine of 9.8; every
  `border.style`; every `border.color`; both bar styles; every palette
  style; icons off; footer off; root user; a custom group layout; a
  layout with every opt-in item.
- Responsive: for one fixture, render at widths 200, 120, 100, 80, 60,
  40, 20, 10, 1 and 0 and unbounded. Assert the width contract at every
  width and assert which layout step was chosen at each.
- Alignment: in every golden output, all lines of the card have equal
  width.
- State handling per item: `Present`, `Absent` and `Unknown`.
- Formatting tables: units, numbers, percent clamping, bar fill at 0, 1,
  50, 99 and 100 percent, uptime, name cleaning with real model strings.
- Hostile input: facts containing escape sequences, newlines, very long
  strings and wide characters do not break the frame.
- Every icon is a single rune in a private-use range.

### 13.14 Commits

1. `Add the theme`
2. `Add the item registry with the default items`
3. `Add rows with aligned labels`
4. `Add usage bars`
5. `Add badges and muted details`
6. `Draw the card with Sections`
7. `Add the palette styles`
8. `Draw distribution logos`
9. `Add gradients for the logo and the frame`
10. `Add the footer`
11. `Add the responsive screen layout`
12. `Add icons`
13. `Add the opt-in items`

---

## 14. Milestone 8: command line

### 14.1 Flags

| Flag | Effect |
|------|--------|
| `--config <path>` | read this configuration file |
| `--logo <name>` | overrides `logo.name` |
| `--no-icons` | overrides `icons.enabled` to false |
| `--no-color` | overrides `colors.mode` to `none` |
| `--clear` | overrides `clear` to true |
| `--print-config` | print the default configuration and exit |
| `--version` | print the version and exit |
| `--help` | print usage and exit |
| `--dont-clear` | accepted and ignored; not listed in the help |

- `--dont-clear` stays so that existing shell rc files keep working after
  the default changed to not clearing.
- `--logo` with an unknown name is a flag error (exit code 2) that lists
  the known names.
- Positional arguments are a flag error.
- `--print-config`, `--version` and `--help` do not read the
  configuration file, so they work when it is broken.

### 14.2 Run

1. Parse flags. On error: message and usage to stderr, exit 2.
2. Load the configuration and apply flag overrides. On error: the
   problems to stderr, exit 2.
3. Compute the facts the configured items need and collect them,
   measuring the time.
4. Build the screen.
5. Create the `beautify.Composer` for stdout. `colors.mode = "auto"`
   lets beautify detect the profile, which also honors `NO_COLOR` and
   disables styling when stdout is not a terminal. Other values fix the
   profile.
6. If `clear` is set and stdout is a terminal, write `ESC [H ESC [2J
   ESC [3J` first. Never clear when stdout is not a terminal. Do not run
   the `clear` program.
7. Render. A write error: message to stderr, exit 1.

### 14.3 Version

`--version` prints `fetchify <version>`. The version comes from
`runtime/debug.ReadBuildInfo` (the module version set by `go install`),
and is `dev` for local builds. No version constant is kept in the source
and no version appears in any logo.

### 14.4 Tests

- Every flag, unknown flags, positional arguments, flag and file
  precedence.
- `--help`, `--version` and `--print-config` succeed with a broken
  configuration file.
- `run` with fake stdout that is not a terminal: no escape sequences, no
  clearing, even with `--clear`.
- A failing writer gives exit code 1.
- The output of `--print-config`, fed back through `--config`, produces
  the same screen as no configuration.

### 14.5 Commits

1. `Move flag parsing to internal/cli`
2. `Load configuration at startup`
3. `Collect only the configured facts`
4. `Detect the color profile and honor NO_COLOR`
5. `Stop clearing the screen by default`
6. `Add --version`

---

## 15. Milestone 9: documentation and release

- Rewrite `README.md`: a screenshot or text sample, installation with
  `go install`, the Nerd Font requirement and how to turn icons off, the
  flags, the configuration file with every key, the list of items marked
  default or opt-in, the list of logos, the trademark note, a link to
  `goals.md`.
- Update `AGENTS.md`: architecture, layers, how to add an item, how to
  add a distribution.
- Delete `install.sh`.
- Change the module path to `go.xchunk.org/fetchify/v2`. Go requires the
  `/v2` suffix for a `v2.0.0` tag; without it the tag is ignored by
  `go install`. The install command becomes
  `go install go.xchunk.org/fetchify/v2@latest`.
- Tag `v2.0.0`. The old `-generic` tag suffix is dropped: in semantic
  versioning it marks a pre-release.

Commits:

1. `Document the architecture in AGENTS.md`
2. `Rewrite README for 2.0`
3. `Remove install.sh`
4. `Change the module path to go.xchunk.org/fetchify/v2`

---

## 16. Cross-cutting edge-case matrix

Each row must be covered by at least one test in every layer it applies
to.

| Case | sysinfo | distro | config | ui | cli |
|------|---------|--------|--------|----|-----|
| Fact absent | x |   |   | x |   |
| Fact unknown or timed out | x |   |   | x |   |
| Collector panics | x |   |   |   |   |
| Tool not installed | x |   |   | x |   |
| Container: no DMI, no DRM, overlay root | x |   |   | x |   |
| Headless or SSH session | x |   |   | x |   |
| Running as root | x |   |   | x |   |
| Unknown distribution | x | x |   | x |   |
| Derivative known only through `ID_LIKE` |   | x |   | x |   |
| Missing or empty os-release | x | x |   | x |   |
| Escape sequences in collected text | x |   |   | x |   |
| Very long values |   |   |   | x |   |
| Wide and combining characters in values |   |   |   | x |   |
| Several GPUs, disks, batteries, displays | x |   |   | x |   |
| No configuration file |   |   | x |   | x |
| Invalid configuration |   |   | x |   | x |
| Empty group, no groups |   |   | x | x |   |
| Every opt-in item enabled | x |   | x | x |   |
| Width 0, 1 and unbounded |   |   |   | x |   |
| Output not a terminal |   |   |   | x | x |
| `NO_COLOR` set |   |   |   |   | x |
| Every color profile |   |   |   | x |   |
| Icons disabled |   |   | x | x | x |

---

## 17. Final verification checklist

- [ ] `gofmt -l .` prints nothing.
- [ ] `go vet ./...` is clean.
- [ ] `go test -race ./...` passes.
- [ ] `go test -run=^$ -fuzz=FuzzLoad -fuzztime=30s ./internal/config/`
      passes.
- [ ] `go.mod` requires only beautify and the TOML library.
- [ ] No package other than `ui` and `main` imports beautify, except
      `distro` for colors if that option was kept.
- [ ] No escape sequence is written anywhere in fetchify except the
      screen clear in `main`.
- [ ] Every exported identifier has a doc comment.
- [ ] Every row of section 16 is covered.
- [ ] The default run takes under 50 ms wall time on a desktop, measured
      with a warm cache.
- [ ] Manual check in a real terminal: default look; a width of 80, 60
      and 40 columns; `--no-icons`; `--no-color`; output piped to a file
      contains no escape sequences; each border style and color mode.
- [ ] Manual check on at least three different distributions, in
      containers if needed: the right logo and package manager.
- [ ] `go install go.xchunk.org/fetchify/v2@latest` installs the tagged
      version and `fetchify --version` prints it.
- [ ] `git log` shows atomic commits that each build and pass tests.
- [ ] Every difference from the beautify plan found while implementing is
      recorded in section 18.

---

## 18. Decision log

Decisions already made. Do not revisit them without the maintainer.

| Decision | Reason |
|----------|--------|
| fetchify 2.0 is implemented after beautify v1 | The look is built from beautify blocks. |
| Card with sections whose dividers join the frame | Chosen look. |
| `user@host` is the caption at the right of the top edge | The first section title occupies the left. |
| Icons are on by default | Maintainer's choice; they can be turned off in the configuration or with `--no-icons`. |
| The palette is eight dots in the bottom edge | Nothing hangs under the card. |
| Level colors and the palette use the 16-color palette; accent, logo and gradients are 24-bit | The first group should follow the terminal theme, the second carries the identity of the distribution. |
| The root warning text is replaced by a red caption | It does not fit in the top edge and the color says the same. |
| Configuration is TOML, validated strictly | A typo is reported instead of silently ignored. |
| Defining any group replaces the default groups | The only way to remove and reorder them predictably. |
| Items that run user commands are not part of this version | Running arbitrary commands at every shell start needs its own design. |
| Named themes are not part of this version | The schema should settle first. |
| The screen is not cleared by default | Clearing destroys what was on screen in manual runs. |
| The version is not shown on screen | It is available through `--version`. |
| Linux only | See `goals.md`. |
| Installation is `go install` only; `install.sh` is removed | One supported way. |
| `gopsutil` is removed | Every fact has a direct Linux source; fewer dependencies and a faster start. |
| `/mnt` is no longer ignored for disks | External drives are worth showing. |
| The license is MIT | Maintainer's choice; it also matches the sources of the logos. |
| The release is `v2.0.0` with the `/v2` module path | New look and changed defaults; the path suffix is required by Go. |

Record additions below this line while implementing: the date, what was
added or changed, and why.
