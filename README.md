# recalbox-ledd

Go daemon for Recalbox that listens to EmulationStation MQTT events and drives NeoPixel ring effects.

## What it does

- Subscribes to `Recalbox/EmulationStation/Event`
- Reads extra context from `/tmp/es_state.inf`
- Maps systems to theme colors based on the frontend theme files
- Runs animated ring effects based on frontend events and current system
- Normalizes dim theme colors so they stay visible on the ring
- Uses a bright default theme when the system is unknown or missing
- Designed to be launched by a permanent Recalbox user script

## Recalbox layout

Copy files to:

- Binary: `/recalbox/share/userscripts/recalbox-ledd/recalbox-ledd`
- Wrapper: `/recalbox/share/userscripts/led-listener(permanent).sh`

The wrapper script copies the binary to `/tmp/recalbox-ledd`, runs `chmod +x` on that staged copy, and executes it from `/tmp`. This avoids execution failures on filesystems where the original deployed binary cannot be marked executable.

## Build notes

This program uses:

- `github.com/eclipse/paho.mqtt.golang`
- `github.com/rpi-ws281x/rpi-ws281x-go`

The ws281x Go package depends on the native `rpi_ws281x` C library, so the final binary is dynamically linked.

## Cross compile for Raspberry Pi Zero 2 W

The Pi Zero 2 W can run either a 64-bit ARM64 image or a 32-bit ARMv7 image. Build the binary that matches your target OS:

```bash
# 64-bit Pi OS/Recalbox
./build-pi-zero-2w.sh linux/arm64

# 32-bit Pi OS/Recalbox
./build-pi-zero-2w.sh linux/arm/v7
```

The output binary is written to:

- `dist/linux-arm64/recalbox-ledd` for 64-bit
- `dist/linux-armv7/recalbox-ledd` for 32-bit

The build uses Docker Buildx because `github.com/rpi-ws281x/rpi-ws281x-go` requires cgo and the native `rpi_ws281x` C library.

## GitHub release build

Pushing a tag that starts with `v` triggers GitHub Actions to build both supported Raspberry Pi binaries and attach them to the corresponding GitHub Release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Release assets:

- `recalbox-ledd-linux-arm64`
- `recalbox-ledd-linux-armv7`

## Native Pi build

Build on a compatible Raspberry Pi / Recalbox-like ARM environment with the native ws281x library installed:

```bash
go build -o recalbox-ledd .
```

## Example deployment

```bash
mkdir -p /recalbox/share/userscripts/recalbox-ledd
cp recalbox-ledd /recalbox/share/userscripts/recalbox-ledd/
cp led-listener\(permanent\).sh /recalbox/share/userscripts/
chmod +x /recalbox/share/userscripts/led-listener\(permanent\).sh
```

## Parameters

- `-mqtt` MQTT broker URL, default `tcp://127.0.0.1:1883`
- `-topic` MQTT topic, default `Recalbox/EmulationStation/Event`
- `-state` ES state file, default `/tmp/es_state.inf`
- `-count` number of LEDs
- `-gpio` GPIO pin, usually `18`
- `-brightness` brightness `0-255`
- `-strip` `ws2812` or `sk6812rgbw`

## Event mapping

- `start`, `wakeup` -> theater chase using the current system theme
- `sleep`, `stop`, `shutdown` -> fade to black
- `systembrowsing` -> snail/comet effect using primary and accent theme colors
- `gamelistbrowsing` -> pulsing system theme
- `rungame` -> rainbow effect tinted by the current system theme
- `endgame` -> themed pulse
- `configurationchanged` -> short accent-color blink

## System colors

- Known systems use colors derived from the EmulationStation theme definitions.
- Some systems also have explicit accent colors for two-tone effects.
- Very dark colors are brightened to a minimum visible level so the ring does not appear off.
- Unknown systems fall back to a bright default blue and amber theme.

## Caveats

- Needs root-level hardware access, which Recalbox scripts typically have.
- The deployed binary itself does not need the executable bit if the wrapper can copy it into `/tmp` and `chmod +x` the staged copy.
- Requires the native `rpi_ws281x` library and matching runtime linker support on the target system.
- Theme colors are normalized for LED visibility, so the ring may appear brighter than the original UI background color values.
