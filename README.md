# LabwcSetup

LabwcSetup is a terminal based assistant for setting up a Labwc desktop session on FreeBSD and GhostBSD. It follows the same general idea as NiriSetup, but replaces the compositor specific package list and configuration model with a Labwc oriented one.

## Base used for the menu

This project's `menu.xml` is adapted from the Joborun style Labwc menu structure you provided, but cleaned up for FreeBSD and GhostBSD use. The general structure is preserved:

* `client-menu`
* `root-menu`
* a nested configuration submenu
* a power submenu

The Linux centric parts were removed or replaced:

* Joborun branding and icon paths were removed
* `gmrun` was replaced with `wofi`
* `leafpad` usage was normalized to `geany`
* Conky specific items were removed
* distro specific web links were removed
* power actions were changed to FreeBSD oriented commands

## Package stack

This build targets the following stack:

* Bar: `sfwbar`
* Launcher: `wofi`
* Terminal: `foot`
* File manager: `pcmanfm`
* Text editor: `geany`
* Browser: `librewolf`
* Extras: `grim`, `slurp`, `swaybg`, `swayidle`, `swaylock`, `wlopm`, `mako`

The default editor in this build is `geany`. If you prefer `leafpad`, replace the `defaultEditor` constant in `LabwcSetup.go` before building.

## What is reused conceptually from NiriSetup

Most of the overall flow can be reused:

* package installation through `pkg`
* system setup for `dbus`, `seatd`, `pam_xdg`, `drm`, and `video` group membership
* runtime directory preparation for Wayland sessions
* log saving

The main compositor specific change is configuration deployment. Niri used a single `config.kdl` file. Labwc uses a config directory under `~/.config/labwc/`.

## Installed Labwc configuration

This project installs the following files into `~/.config/labwc/`:

* `rc.xml`
* `menu.xml`
* `autostart`
* `environment`
* `backgrounds/dark_blue_bg.png`

## Behavior of the shipped configuration

### Key bindings

* `W-Return` launches `foot`
* `W-space` launches `wofi --show drun`
* `W-e` launches `pcmanfm`
* `W-b` launches `librewolf`
* `W-t` launches `geany`
* `W-l` launches `swaylock`

### Autostart

The autostart file does the following:

* starts `mako`
* starts `sfwbar`
* starts `swaybg` with the bundled dark blue background
* starts `swayidle`
* locks the screen after 5 minutes
* powers displays off after another 5 minutes with `wlopm`
* restores displays on resume

### Root menu

The root menu includes:

* browser
* terminal
* launcher
* file manager
* text editor
* screenshot capture
* lock screen
* bar start and stop
* configuration editing entries
* reconfigure action
* power submenu for `shutdown -p now`, `reboot`, and `zzz`

## Build

Install Go if it is not already present:

```sh
sudo pkg install go
```

Build LabwcSetup:

```sh
go mod tidy
go build -o LabwcSetup .
```

## Run

```sh
./LabwcSetup
```

## Starting Labwc manually

After package installation and system setup, you can start Labwc from a TTY with:

```sh
LIBSEAT_BACKEND=consolekit2 ck-launch-session dbus-launch labwc
```

## Notes for GhostBSD

GhostBSD may already include some of the supporting Wayland packages and services. That is expected. The setup step may therefore report that some components are already present.

The power actions in `menu.xml` assume a privilege model that allows the chosen commands to run. If your GhostBSD setup uses `sudo`, `doas`, or another helper, adjust those menu entries to match local policy.

## Menu customization notes

If you want to switch from `geany` to `leafpad`, update both:

* the `defaultEditor` constant in `LabwcSetup.go`
* the `Text Editor` and `edit ...` entries in `configs/labwc/menu.xml`

## Logs

Logs are written to:

```text
/tmp/labwcsetup.log
```
