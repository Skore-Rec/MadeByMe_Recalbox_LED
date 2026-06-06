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

After Recalbox is installed, the SD card exposes a `SHARE` partition when mounted on another computer. That partition maps to `/recalbox/share` on the running system, so this project should be copied into `SHARE/userscripts/` on the card.

The wrapper script copies the binary to `/tmp/recalbox-ledd`, runs `chmod +x` on that staged copy, and executes it from `/tmp`. This avoids execution failures on filesystems where the original deployed binary cannot be marked executable.

## Quick start

If you just want to try it, use the prebuilt files from the latest GitHub Release.

Download:

- `recalbox-ledd-linux-arm64` for 64-bit Recalbox
- `recalbox-ledd-linux-armv7` for 32-bit Recalbox
- `led-listener(permanent).sh` from this repository

Then:

1. Rename the downloaded release file to `recalbox-ledd`.
2. Shut down the Pi and put the SD card into your computer.
3. Open the mounted `SHARE` partition.
4. Create `userscripts/recalbox-ledd/` inside `SHARE`.
5. Copy `recalbox-ledd` into `SHARE/userscripts/recalbox-ledd/`.
6. Copy `led-listener(permanent).sh` into `SHARE/userscripts/`.
7. Put the SD card back into the Pi and boot Recalbox.

On the running system, those files will appear at `/recalbox/share/userscripts/recalbox-ledd/recalbox-ledd` and `/recalbox/share/userscripts/led-listener(permanent).sh`.

## Example deployment

If you are copying files directly onto the Recalbox SD card:

1. Shut down the Pi and insert the SD card into your computer.
2. Open the mounted `SHARE` partition.
3. Create `userscripts/recalbox-ledd/` inside `SHARE` if it does not already exist.
4. Copy the `recalbox-ledd` binary into `SHARE/userscripts/recalbox-ledd/`.
5. Copy `led-listener(permanent).sh` into `SHARE/userscripts/`.
6. Put the SD card back into the Pi and boot Recalbox.

On macOS, the mounted path is usually `/Volumes/SHARE`, so the copy commands look like this:

```bash
mkdir -p /Volumes/SHARE/userscripts/recalbox-ledd
cp recalbox-ledd /Volumes/SHARE/userscripts/recalbox-ledd/
cp led-listener\(permanent\).sh /Volumes/SHARE/userscripts/
```

From Recalbox's point of view, those files will then appear at `/recalbox/share/userscripts/recalbox-ledd/recalbox-ledd` and `/recalbox/share/userscripts/led-listener(permanent).sh`.

```bash
mkdir -p /recalbox/share/userscripts/recalbox-ledd
cp recalbox-ledd /recalbox/share/userscripts/recalbox-ledd/
cp led-listener\(permanent\).sh /recalbox/share/userscripts/
chmod +x /recalbox/share/userscripts/led-listener\(permanent\).sh
```

## Build from source

Only use this section if you want to build the binary yourself.

### Cross compile for Raspberry Pi Zero 2 W

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

### Build release files

Pushing a tag that starts with `v` triggers GitHub Actions to build both supported Raspberry Pi binaries and attach them to the matching GitHub Release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Release assets:

- `recalbox-ledd-linux-arm64`
- `recalbox-ledd-linux-armv7`

### Native Pi build

Build on a compatible Raspberry Pi / Recalbox-like ARM environment with the required LED library installed:

```bash
go build -o recalbox-ledd .
```

## Hardware wiring

This daemon expects the LED data line on a PWM-capable Raspberry Pi pin. The default and most common choice is GPIO `18`.

Minimum direct wiring for a small ring or a few pixels:

- NeoPixel `DIN` -> Pi GPIO `18` (physical pin `12`)
- NeoPixel `GND` -> Pi `GND` (for example physical pin `6`)
- NeoPixel `5V` -> Pi `5V` (physical pin `2` or `4`) only if you are powering just a few LEDs

Recommended wiring for a reliable install:

- Pi GPIO `18` -> `74AHCT125` input
- `74AHCT125` output -> NeoPixel `DIN`
- External `5V` power supply -> NeoPixel `5V`
- External power supply `GND` -> NeoPixel `GND`
- Pi `GND` -> the same shared ground as the power supply and LEDs

Notes:

- Connect to the strip or ring `DIN` pad, not `DOUT`.
- For more than a few LEDs, do not power the strip from the Pi's `5V` pin. Use a separate `5V` supply sized for the LED count.
- The grounds must be common between the Pi and the LED power supply or the data signal will not be referenced correctly.
- A `74AHCT125` level shifter is the preferred way to convert the Pi's `3.3V` data signal up to `5V`. Direct connection sometimes works, but it is not guaranteed.
- Some libraries also support GPIO `10`, `12`, and `21`, but this project defaults to GPIO `18` and its `-gpio` flag should match the hardware wiring.

Reference wiring guide: <https://learn.adafruit.com/neopixels-on-raspberry-pi/raspberry-pi-wiring>

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
- GPIO `18` is shared with onboard audio on Raspberry Pi systems. To use GPIO `18` for the LED strip, edit `/boot/config.txt`, change `dtparam=audio=on` to `dtparam=audio=off`, and reboot. See the Adafruit setup notes: <https://learn.adafruit.com/neopixels-on-raspberry-pi>.
- The deployed binary itself does not need the executable bit if the wrapper can copy it into `/tmp` and `chmod +x` the staged copy.
- Requires the native `rpi_ws281x` library and matching runtime linker support on the target system.
- Theme colors are normalized for LED visibility, so the ring may appear brighter than the original UI background color values.
