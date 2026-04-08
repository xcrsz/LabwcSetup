# LabwcSetup

LabwcSetup is a minimal setup utility for configuring a Labwc-based Wayland desktop on FreeBSD and GhostBSD.

It installs required packages, prepares the system environment, and deploys a working Labwc configuration using a simple, reproducible layout.

This project intentionally avoids unnecessary components and focuses on a clean, functional compositor stack.

---

## Profiles

LabwcSetup supports two installation profiles:

### Minimal Profile

A lightweight, essential environment:

* labwc
* foot (terminal)
* wofi (launcher)

This profile provides a clean base for custom setups.

---

### Extended Profile (default)

A complete, ready-to-use desktop:

* **Compositor**: labwc
* **Bar**: waybar
* **Launcher**: wofi
* **Terminal**: foot
* **File Manager**: pcmanfm
* **Text Editor**: geany
* **Browser**: librewolf
* **Notifications**: mako
* **Lock / Idle**: swayidle + swaylock
* **Background**: swaybg
* **Screenshots**: grim + slurp
* **Audio control**: pavucontrol

---

## Installation

Clone the repository:

```
git clone <repo>
cd LabwcSetup
```

Build the tool:

```
go build -o labwcsetup
```

Run:

```
./labwcsetup
```

You will be prompted to select:

* Minimal
* Extended (default)

---

## System Requirements

Ensure the following base services are enabled and running:

```
sysrc dbus_enable=YES
sysrc seatd_enable=YES

service dbus start
service seatd start
```

---

## User Configuration

Add your user to the video group:

```
pw groupmod video -m $USER
```

Enable `pam_xdg`:

Edit:

```
/etc/pam.d/system
```

Uncomment:

```
session optional pam_xdg.so
```

---

## Running Labwc (GhostBSD / FreeBSD)

Once configured, Labwc can be started directly from a TTY.

### Start Labwc

```
labwc
```

No `sudo` is required.

---

## Configuration

Labwc configuration is installed to:

```
~/.config/labwc/
```

This includes:

* `rc.xml`
* `menu.xml`
* `autostart`
* `environment`
* `backgrounds/dark_blue_bg.png`

If configuration is missing:

```
cp -r configs/labwc ~/.config/
```

---

## Autostart Behavior

### Minimal Profile

Starts only the compositor.

No background services are launched.

---

### Extended Profile

The default session starts:

* waybar
* mako
* swaybg (dark blue background)
* swayidle (screen lock and display power management)

---

## Menu

The right-click root menu provides:

* Application launchers (terminal, browser, file manager, editor)
* Wofi launcher
* Waybar control
* Screenshot tools
* Labwc configuration editing
* Exit Labwc

Notes:

* Power management entries are intentionally removed for GhostBSD compatibility
* Commands are executed via `sh -lc` for correct environment handling

---

## Screenshots

Region screenshot:

```
grim -g "$(slurp)" ~/Pictures/Screenshots/screenshot.png
```

---

## Notes

* `consolekit2` is not required and is intentionally omitted
* If `XDG_RUNTIME_DIR` errors appear, verify `pam_xdg` is enabled
* If the compositor fails to start, ensure:

  * `seatd` is running
  * user is in `video` group
* Configuration must exist in `~/.config/labwc/` or Labwc will fall back to defaults

---

## Design Philosophy

This setup follows a composable system model:

* minimal base (mechanism)
* optional extensions (policy)
* explicit configuration

This reduces complexity and cognitive load while maintaining flexibility.

---

## Future Direction

Planned improvements:

* additional profiles (developer, minimal GUI, etc.)
* hardware validation integration
* tighter GhostBSD integration
* reproducible configuration bundles

---

## License

BSD 2-Clause
