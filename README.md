# LabwcSetup

LabwcSetup is a terminal based assistant for setting up a Labwc desktop session on FreeBSD and GhostBSD. It follows the same general idea as NiriSetup, but replaces the compositor specific package list and configuration model with a Labwc oriented one.

## What this variant changes

This version swaps the original Niri focused components for the following Labwc oriented stack:

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

## Labwc configuration files used here

This project installs the following starter files into `~/.config/labwc/`:

* `rc.xml`
* `autostart`
* `environment`
* `backgrounds/dark_blue_bg.png`

This matches Labwc's documented configuration model. Labwc can also use `menu.xml`, `shutdown`, and `xinitrc`, but they are not required for this starter build.

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


The autostart file now also includes a background setup and idle management block:

* `swaybg` starts with the included `backgrounds/dark_blue_bg.png` wallpaper by default
* `swayidle` locks the screen after 5 minutes
* `wlopm` turns displays off after another 5 minutes and restores them on resume
* `swaylock` is used for both idle locking and before sleep

Adjust the `swaybg` line in `configs/labwc/autostart` if you want to use a different wallpaper or switch back to a solid color.

## Menu actions

LabwcSetup provides these actions:

1. `Install Labwc`
2. `Setup System`
3. `Install Default Config`
4. `Verify Setup`
5. `Save Logs`
6. `Exit`

## Starting Labwc manually

After package installation and system setup, you can start Labwc from a TTY with:

```sh
LIBSEAT_BACKEND=consolekit2 ck-launch-session dbus-launch labwc
```

## Notes for GhostBSD

GhostBSD already carries some of the base Wayland related pieces, so the package installation step may report that some items are already installed. That is expected.

## Config bundle summary

The shipped config does the following:

* `W-Return` launches `foot`
* `W-space` launches `wofi --show drun`
* `W-e` launches `pcmanfm`
* `W-b` launches `librewolf`
* `W-t` launches `geany`
* `W-l` launches `swaylock`
* `autostart` starts `mako`, `sfwbar`, `swaybg`, and `swayidle` using the included dark blue wallpaper

## Logs

Logs are written to:

```text
/tmp/labwcsetup.log
```
