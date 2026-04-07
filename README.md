# LabwcSetup

LabwcSetup is a terminal based assistant for setting up a Labwc desktop session on FreeBSD and GhostBSD. It started from the NiriSetup model, but is adapted for Labwc and for GhostBSD specific behavior reported during testing.

## Current package stack

This build now targets the following stack:

* Bar: `waybar`
* Audio control helper: `pavucontrol`
* Launcher: `wofi`
* Terminal: `foot`
* File manager: `pcmanfm`
* Text editor: `geany`
* Browser: `librewolf`
* Notifications: `mako`
* Extras: `grim`, `slurp`, `swaybg`, `swayidle`, `swaylock`, `wlopm`

## Important GhostBSD and FreeBSD notes

Testing found four practical requirements:

1. add the user to the `video` group
2. uncomment `pam_xdg` in `/etc/pam.d/system`
3. copy the Labwc config directory into `~/.config/labwc`
4. use shell wrapped commands in `menu.xml` so home directory paths expand correctly

The setup step in this project adds the current user to the `video` group and enables the supporting services. After group membership changes, log out and back in.

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
* starts `waybar`
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
* audio control with `pavucontrol`
* screenshot capture
* lock screen
* bar start and stop
* configuration editing entries
* reconfigure action
* exit Labwc

The power submenu was removed because GhostBSD does not assume passwordless `sudo`, and the earlier `zzz` based entry was not appropriate as a default.

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
ck-launch-session dbus-launch labwc
```

If `pam_xdg` is enabled and the user is in the `video` group, Labwc should start without `sudo`.

## Logs

Logs are written to:

```text
/tmp/labwcsetup.log
```
