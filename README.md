# fetchify

A fetch tool for Linux: it prints facts about the machine in a framed
card next to the logo of its distribution.

```
                                          ╭─ * System ─────────────────────────────────── alice@box ─╮
             --+oossssssoo+--             │ * OS        Ubuntu 24.04 LTS                             │
         .:+ssssssssssssssssss+:.         │ * Kernel    6.8.0-45-generic                             │
       -+ssssssssssssssssssyyssss+-       │ * Uptime    3h 25m                                       │
     .ossssssssssssssssssdMMMNysssso.     │ * Packages  1842 (dpkg), 12 (flatpak)                    │
   /ssssssssssshdmmNNmmyNMMMMhssssss/     │ * Shell     zsh                                          │
  +ssssssssshmydMMMMMMMNddddyssssssss+    │ * Terminal  kitty                                        │
 /sssssssshNMMMyhhyyyyhmNMMMNhssssssss/   │ * DE        GNOME                                        │
.ssssssssdMMMNhsssssssssshNMMMdssssssss.  │ * WM        Mutter  Wayland                              │
+sssshhhyNMMNyssssssssssssyNMMMysssssss+  ├─ * Hardware ─────────────────────────────────────────────┤
ossyNMMMNyMMhsssssssssssssshmmmhssssssso  │ * Host      ThinkPad X1 Carbon Gen 11                    │
ossyNMMMNyMMhsssssssssssssshmmmhssssssso  │ * CPU       AMD Ryzen 7 5800X  16 cores                  │
+sssshhhyNMMNyssssssssssssyNMMMysssssss+  │ * GPU       NVIDIA GeForce RTX 3070                      │
.ssssssssdMMMNhsssssssssshNMMMdssssssss.  │ * Display   2560x1440                                    │
 /sssssssshNMMMyhhyyyyhdNMMMNhssssssss/   │ * Battery   ████████████░░░░  78%  charging              │
  +sssssssssdmydMMMMMMMMddddyssssssss+    ├─ * Storage ──────────────────────────────────────────────┤
   /ssssssssssshdmNNNNmyNMMMMhssssss/     │ * RAM       ████░░░░░░░░░░░░  8.0 / 32.0 GiB  25%        │
    .ossssssssssssssssssdMMMNysssso.      │ * Swap      ████░░░░░░░░░░░░  1.0 /  4.0 GiB  25%        │
     -+sssssssssssssssssyyyssss+-         │ * /         ███░░░░░░░░░░░░░  100 /  500 GiB  21%  ext4  │
       `:+ssssssssssssssssss+:´           │ * /home     ████░░░░░░░░░░░░  250 / 1000 GiB  25%  btrfs │
           --+oossssssoo+--               ╰──────────────────────────────────────── ● ● ● ● ● ● ● ● ─╯
                                            collected in 12 ms
```

Icons are shown as `*` here; a terminal with a Nerd Font shows glyphs.
The card adapts to the width of the terminal: the logo shrinks, then
disappears, then details, bars and finally the ends of values go.

## Installation

```sh
go install go.xchunk.org/fetchify/v2@latest
```

This needs Go 1.24 or newer and puts `fetchify` into `$(go env GOPATH)/bin`.
To show it in every new shell, add `fetchify` to the end of your
`~/.bashrc`, `~/.zshrc` or `~/.config/fish/config.fish`.

### Icons

Icons need a [Nerd Font](https://www.nerdfonts.com) in the terminal.
Without one, turn them off with `--no-icons` or in the configuration:

```toml
[icons]
enabled = false
```

## Command line

| Flag | Effect |
|------|--------|
| `--config <path>` | read this configuration file |
| `--logo <name>` | show this logo: `auto`, `none`, `koala` or a distribution |
| `--no-icons` | do not show icons |
| `--no-color` | do not use colors; `NO_COLOR` has the same effect |
| `--clear` | clear the screen before printing |
| `--print-config` | print the documented default configuration and exit |
| `--version` | print the version and exit |
| `--help` | print the usage and exit |

Flags win over the configuration file. Output that does not go to a
terminal is plain text without escape sequences. Invalid flags and an
invalid configuration print every problem and exit with code 2.

## Configuration

fetchify reads `$XDG_CONFIG_HOME/fetchify/config.toml`, or
`~/.config/fetchify/config.toml`. The file is optional; keys that are
left out keep their defaults, and unknown keys are errors. This is the
complete default configuration, as `fetchify --print-config` prints it:

```toml
# fetchify configuration
#
# Location: $XDG_CONFIG_HOME/fetchify/config.toml, or
# ~/.config/fetchify/config.toml. Keys left out keep the values shown
# here, and an array replaces its default as a whole.

clear      = false     # clear the screen before printing
footer     = true      # the line under the card
timeout_ms = 200       # default time limit per item, 1-10000

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
width      = 16        # 4-40
thresholds = [60, 85]  # percent where the color turns to warning and danger

[palette]
style = "dots"         # dots | blocks | none

# Groups are the sections of the card, in order. Defining any group
# replaces all three default groups. icon = "auto" takes the icon of the
# default group with the same title.

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

# Settings of single items. Every item accepts:
#   label      = "..."   replaces the label
#   icon       = "..."   replaces the icon; "" removes it
#   timeout_ms = 500     replaces the time limit
# Item-specific keys: items.disks.ignore, items.memory.show_type,
# items.public_ip.url.

[items.disks]
ignore = ["/boot", "/efi", "/etc", "/run", "/var", "/tmp", "/dev", "/sys", "/proc", "/snap", "/nix/store"]
```

Defining any `[[group]]` replaces the three default groups, so a file
can remove and reorder them. A group whose items have nothing to show
is not drawn.

## Items

Default items:

| Item | Shows |
|------|-------|
| `os` | name and version of the distribution |
| `kernel` | kernel release |
| `uptime` | time since boot |
| `packages` | installed packages per package manager |
| `shell` | login shell |
| `terminal` | terminal emulator |
| `de` | desktop environment, hidden when it equals the window manager |
| `wm` | window manager and the session type |
| `host` | model of the machine |
| `cpu` | processor and its number of logical CPUs |
| `gpu` | one row per GPU |
| `display` | resolution of the connected displays |
| `battery` | charge and state, one row per battery |
| `memory` | RAM use; `items.memory.show_type = true` adds the module types (needs root) |
| `swap` | swap use |
| `disks` | one row per file system, labeled with its mount point |

Opt-in items, collected only when a group lists them:

| Item | Shows | Fits into |
|------|-------|-----------|
| `local_ip` | first global address and its interface | System |
| `public_ip` | address on the internet, asked from `items.public_ip.url` | System |
| `network` | network of the default route | System |
| `bluetooth` | whether Bluetooth is on | Hardware |
| `cpu_temp` | processor temperature | Hardware |
| `gpu_temp` | temperature per GPU | Hardware |
| `cpu_usage` | processor usage, sampled over 100 ms | Hardware |
| `load` | load average | System |
| `gpu_driver` | kernel driver per GPU | Hardware |
| `gpu_memory` | video memory use per GPU | Storage |
| `board` | main board | Hardware |
| `bios` | firmware vendor, version and date | Hardware |
| `init` | init system | System |
| `processes` | number of processes | System |
| `users` | number of logged-in users | System |
| `locale` | locale of messages | System |
| `timezone` | time zone | System |
| `date` | date and time | System |
| `theme`, `icon_theme`, `font`, `cursor` | look of the desktop | System |
| `terminal_font` | font of kitty, Alacritty, foot, WezTerm or Ghostty | System |
| `media` | track of a media player, through `playerctl` | System |

An item whose fact does not exist on the machine is hidden; one that
could not be determined shows `unknown`.

## Logos

`almalinux`, `alpine`, `arch`, `artix`, `cachyos`, `centos`, `debian`, `elementary`, `endeavouros`, `fedora`, `gentoo`, `kali`, `koala`, `linux`, `linuxmint`, `manjaro`, `nixos`, `opensuse`, `pop`, `rhel`, `rocky`, `ubuntu`, `void`, `zorin`

An unknown distribution gets the logo of the first known one in its
`ID_LIKE`, and otherwise Tux (`linux`). `koala` is the fetchify mascot.

Distribution names and logos are trademarks of their owners and are used
only to identify the running system. Most logos come from
[fastfetch](https://github.com/fastfetch-cli/fastfetch) and
[neofetch](https://github.com/dylanaraps/neofetch); see
`internal/distro/logos/NOTICE`.

## Scope

fetchify supports Linux only. `goals.md` lists what is deliberately left
for later, such as other operating systems, named themes and image
logos.

## License

MIT; see `LICENSE`.
