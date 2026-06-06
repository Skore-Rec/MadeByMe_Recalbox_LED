package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

type Config struct {
	MQTTBroker string
	MQTTTopic  string
	StateFile  string
	LEDCount   int
	GPIOPin    int
	Brightness int
	StripType  string
	LogPrefix  string
}

type App struct {
	cfg    Config
	strip  *ws2811.WS2811
	client mqtt.Client

	mu           sync.Mutex
	effectCancel context.CancelFunc
}

type SystemTheme struct {
	Primary uint32
	Accent  uint32
}

const (
	defaultPrimaryColor = 0x3c78d8
	defaultAccentColor  = 0xffb000
	minThemeBrightness  = 112
)

var systemBaseColors = map[string]uint32{
	"240ptestsuite":  0x0f143a,
	"3do":            0x2c1f5e,
	"3ds":            0x888888,
	"64dd":           0x333333,
	"amiga1200":      0x2e0e0f,
	"amiga600":       0x2e0e0f,
	"amigacd32":      0x910e00,
	"amigacdtv":      0x490700,
	"amstradcpc":     0x7a1837,
	"apple2":         0x494842,
	"apple2gs":       0x6b6b6b,
	"arcade":         0xd148ec,
	"arduboy":        0x7f571f,
	"atari2600":      0x630503,
	"atari5200":      0x0f3160,
	"atari7800":      0xafafaf,
	"atari800":       0xffb600,
	"atarist":        0x1e647f,
	"atomiswave":     0x264f31,
	"bbcmicro":       0x888888,
	"bk":             0x151515,
	"c64":            0x474237,
	"cassettevision": 0x494949,
	"cdi":            0x14617a,
	"channelf":       0x777449,
	"colecovision":   0xcc6651,
	"creativision":   0x494949,
	"daphne":         0x888888,
	"dice":           0x383838,
	"dos":            0x3d003d,
	"dragon":         0x6d8e12,
	"dreamcast":      0x2c436b,
	"easyrpg":        0x34422b,
	"favorites":      0xd49800,
	"fbneo":          0xaa663f,
	"fds":            0xff9000,
	"gamecube":       0x4848a5,
	"gamegear":       0x004e82,
	"gb":             0x6f7c08,
	"gba":            0x14046d,
	"gbc":            0x01756d,
	"gw":             0xcccccc,
	"gx4000":         0xd4a500,
	"imageviewer":    0x008710,
	"intellivision":  0x5e5442,
	"jaguar":         0x6b0a00,
	"lowresnx":       0xdb3d04,
	"lutro":          0x0a042b,
	"lynx":           0xcc351e,
	"macintosh":      0xe5ddb3,
	"mame":           0x182e47,
	"mastersystem":   0x001b42,
	"megadrive":      0x22427c,
	"megaduck":       0xd7d8d7,
	"model3":         0x888888,
	"moonlight":      0x888888,
	"msx1":           0xdbdbdb,
	"msx2":           0x4a0000,
	"msxturbor":      0x919191,
	"multivision":    0xe3cc00,
	"n64":            0x575123,
	"naomi":          0x890000,
	"naomi2":         0x9c3900,
	"naomigd":        0x471e00,
	"nds":            0x888888,
	"neogeo":         0x937632,
	"neogeocd":       0x878787,
	"nes":            0x6b6b6b,
	"ngp":            0x990000,
	"ngpc":           0x18a94e,
	"o2em":           0xd9b657,
	"openbor":        0xe68181,
	"oricatmos":      0xaebaad,
	"p2000t":         0x6b6b6b,
	"palm":           0x465fab,
	"pc88":           0x888888,
	"pc98":           0x888888,
	"pcengine":       0x520101,
	"pcenginecd":     0x610d0d,
	"pcfx":           0x39255e,
	"pcv2":           0x0c0914,
	"pico":           0x1ab5bb,
	"pico8":          0xf2824b,
	"pokemini":       0xf8e023,
	"ports":          0x4c3856,
	"ps2":            0x6b6b6b,
	"ps3":            0x6b6b6b,
	"psp":            0x6b6b6b,
	"psx":            0x8a8484,
	"samcoupe":       0x96958c,
	"satellaview":    0x647a54,
	"saturn":         0x0e0e1b,
	"scummvm":        0xce6b0a,
	"scv":            0xffb81f,
	"sega32x":        0x5b5b7d,
	"segacd":         0x1c5099,
	"sg1000":         0x134ab7,
	"snes":           0x490000,
	"solarus":        0x9b8eff,
	"spectravideo":   0x4a4237,
	"sufami":         0x4d4248,
	"supergrafx":     0x0d4251,
	"supervision":    0x08d2bf,
	"thomson":        0x0880a1,
	"ti994a":         0x9c9f97,
	"tic80":          0x212333,
	"trs80coco":      0x9a9caf,
	"uzebox":         0x888888,
	"vectrex":        0x07dddf,
	"vg5000":         0xffc000,
	"vic20":          0xb58345,
	"videopacplus":   0x292f33,
	"vircon32":       0x6b6b6b,
	"virtualboy":     0x47060c,
	"vpinball":       0x680f24,
	"wasm4":          0xff52a2,
	"wii":            0x36b8d8,
	"wiiu":           0x9e225b,
	"wswan":          0x9e225b,
	"wswanc":         0x085f87,
	"x1":             0x888888,
	"x68000":         0x0f780f,
	"xbox":           0x0f780f,
	"zmachine":       0x81f2cc,
	"zx81":           0x852828,
	"zxspectrum":     0x888888,
}

var systemAliases = map[string]string{
	"amiga":                            "amiga1200",
	"amigaecsocs":                      "amiga1200",
	"amstradgx4000":                    "gx4000",
	"appleii":                          "apple2",
	"appleiigs":                        "apple2gs",
	"atari8bits":                       "atari800",
	"atarilynx":                        "lynx",
	"commodore64":                      "c64",
	"gameboy":                          "gb",
	"gameboyadvance":                   "gba",
	"gameboycolor":                     "gbc",
	"genesis":                          "megadrive",
	"mattelintellivision":              "intellivision",
	"msx":                              "msx1",
	"neogeoaes":                        "neogeo",
	"nintendoentertainmentsystem":      "nes",
	"philipsvg5000":                    "vg5000",
	"playstation":                      "psx",
	"pokemonmini":                      "pokemini",
	"pokémonmini":                      "pokemini",
	"segagamegear":                     "gamegear",
	"segamastersystemmarkiii":          "mastersystem",
	"segamegadrive":                    "megadrive",
	"segasg1000":                       "sg1000",
	"supernintendo":                    "snes",
	"supernintendoentertainmentsystem": "snes",
}

var systemAccentColors = map[string]uint32{
	"mastersystem": 0x0075ba,
	"saturn":       0x45456b,
	"snes":         0x751515,
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.MQTTBroker, "mqtt", "tcp://127.0.0.1:1883", "MQTT broker URL")
	flag.StringVar(&cfg.MQTTTopic, "topic", "Recalbox/EmulationStation/Event", "MQTT topic")
	flag.StringVar(&cfg.StateFile, "state", "/tmp/es_state.inf", "EmulationStation state file")
	flag.IntVar(&cfg.LEDCount, "count", 12, "LED count")
	flag.IntVar(&cfg.GPIOPin, "gpio", 18, "GPIO pin (typically 18 for PWM)")
	flag.IntVar(&cfg.Brightness, "brightness", 96, "Brightness 0-255")
	flag.StringVar(&cfg.StripType, "strip", "ws2812", "Strip type: ws2812|sk6812rgbw")
	flag.StringVar(&cfg.LogPrefix, "log-prefix", "recalbox-ledd", "Log prefix")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("%s starting", cfg.LogPrefix)

	app, err := newApp(cfg)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer app.close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.run(ctx); err != nil && ctx.Err() == nil {
		log.Fatalf("run failed: %v", err)
	}
}

func newApp(cfg Config) (*App, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = cfg.Brightness
	opt.Channels[0].GpioPin = cfg.GPIOPin
	opt.Channels[0].LedCount = cfg.LEDCount
	opt.Channels[0].Invert = false
	opt.Frequency = 800000
	opt.DmaNum = 10

	if strings.EqualFold(cfg.StripType, "sk6812rgbw") {
		opt.Channels[0].StripeType = ws2811.SK6812StripRGBW
	} else {
		opt.Channels[0].StripeType = ws2811.WS2811StripGRB
	}

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}
	if err := dev.Init(); err != nil {
		return nil, err
	}

	app := &App{cfg: cfg, strip: dev}
	if err := app.setAll(0x000000); err != nil {
		return nil, err
	}
	return app, nil
}

func (a *App) run(ctx context.Context) error {
	opts := mqtt.NewClientOptions().AddBroker(a.cfg.MQTTBroker)
	opts.SetClientID(fmt.Sprintf("recalbox-ledd-%d", time.Now().Unix()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(3 * time.Second)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("mqtt connected, subscribing to %s", a.cfg.MQTTTopic)
		if token := c.Subscribe(a.cfg.MQTTTopic, 0, a.handleMessage); token.Wait() && token.Error() != nil {
			log.Printf("subscribe error: %v", token.Error())
		}
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	})

	a.client = mqtt.NewClient(opts)
	if token := a.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	<-ctx.Done()
	log.Printf("shutdown requested")
	return nil
}

func (a *App) close() {
	a.stopEffect()
	if a.client != nil && a.client.IsConnected() {
		a.client.Disconnect(250)
	}
	if a.strip != nil {
		if err := a.setAll(0x000000); err != nil {
			log.Printf("render error during shutdown: %v", err)
		}
		a.strip.Fini()
	}
}

func (a *App) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	event := strings.ToLower(strings.TrimSpace(string(msg.Payload())))
	state := readStateFile(a.cfg.StateFile)
	system := state["System"]
	game := filepath.Base(state["Game"])
	theme := themeForSystem(system)
	log.Printf("event=%s system=%s game=%s", event, system, game)

	switch event {
	case "start", "wakeup":
		a.startEffect(func(ctx context.Context) {
			a.runTheaterChase(ctx, theme.Primary, theme.Accent, 90*time.Millisecond)
		})
	case "sleep", "stop", "shutdown":
		a.startEffect(func(ctx context.Context) {
			a.runFade(ctx, 0x202020, 0x000000, 18, 35*time.Millisecond)
		})
	case "systembrowsing":
		a.startEffect(func(ctx context.Context) {
			a.runSnail(ctx, theme.Primary, theme.Accent, 80*time.Millisecond)
		})
	case "gamelistbrowsing":
		a.startEffect(func(ctx context.Context) {
			a.runPulse(ctx, theme.Primary, theme.Accent, 40*time.Millisecond)
		})
	case "rungame":
		a.startEffect(func(ctx context.Context) {
			a.runRainbow(ctx, theme.Primary, theme.Accent, 70*time.Millisecond)
		})
	case "endgame":
		a.startEffect(func(ctx context.Context) {
			a.runPulse(ctx, theme.Primary, theme.Accent, 45*time.Millisecond)
		})
	case "configurationchanged":
		a.startEffect(func(ctx context.Context) {
			a.runBlink(ctx, theme.Accent, 4, 120*time.Millisecond, scaleColor(theme.Primary, 56))
		})
	default:
		log.Printf("unhandled event: %s", event)
	}
}

func (a *App) solid(color uint32) {
	a.stopEffect()
	if err := a.setAll(color); err != nil {
		log.Printf("render error: %v", err)
	}
}

func (a *App) stopEffect() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.effectCancel != nil {
		a.effectCancel()
		a.effectCancel = nil
	}
}

func (a *App) startEffect(run func(context.Context)) {
	a.mu.Lock()
	if a.effectCancel != nil {
		a.effectCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.effectCancel = cancel
	a.mu.Unlock()

	go run(ctx)
}

func (a *App) setAll(color uint32) error {
	frame := make([]uint32, a.cfg.LEDCount)
	for i := range frame {
		frame[i] = color
	}
	return a.renderFrame(frame)
}

func (a *App) renderFrame(frame []uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.strip == nil {
		return nil
	}

	leds := a.strip.Leds(0)
	limit := min(len(leds), min(len(frame), a.cfg.LEDCount))
	for i := 0; i < limit; i++ {
		leds[i] = frame[i]
	}
	for i := limit; i < min(len(leds), a.cfg.LEDCount); i++ {
		leds[i] = 0
	}
	return a.strip.Render()
}

func (a *App) runSnail(ctx context.Context, primary uint32, accent uint32, stepDelay time.Duration) {
	ticker := time.NewTicker(stepDelay)
	defer ticker.Stop()

	pos := 0
	tail := []uint8{255, 180, 108, 56}
	for {
		frame := make([]uint32, a.cfg.LEDCount)
		for i := range frame {
			frame[i] = scaleColor(accent, 18)
		}
		for offset, level := range tail {
			idx := (pos - offset + a.cfg.LEDCount) % a.cfg.LEDCount
			base := accent
			if offset < 2 {
				base = primary
			}
			frame[idx] = scaleColor(base, level)
		}
		if err := a.renderFrame(frame); err != nil {
			log.Printf("render error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pos = (pos + 1) % a.cfg.LEDCount
		}
	}
}

func (a *App) runPulse(ctx context.Context, primary uint32, accent uint32, stepDelay time.Duration) {
	ticker := time.NewTicker(stepDelay)
	defer ticker.Stop()

	level := uint8(72)
	delta := 12
	blendStep := 0
	blendDelta := 1
	for {
		color := blendColor(primary, accent, blendStep, 8)
		if err := a.setAll(scaleColor(color, level)); err != nil {
			log.Printf("render error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if int(level)+delta >= 255 || int(level)+delta <= 56 {
				delta = -delta
			}
			level = uint8(int(level) + delta)
			if blendStep+blendDelta >= 8 || blendStep+blendDelta <= 0 {
				blendDelta = -blendDelta
			}
			blendStep += blendDelta
		}
	}
}

func (a *App) runRainbow(ctx context.Context, primary uint32, accent uint32, stepDelay time.Duration) {
	ticker := time.NewTicker(stepDelay)
	defer ticker.Stop()

	shift := 0
	for {
		frame := make([]uint32, a.cfg.LEDCount)
		for i := 0; i < a.cfg.LEDCount; i++ {
			index := (i*256/a.cfg.LEDCount + shift) % 256
			wheel := colorWheel(index)
			tint := blendColor(primary, accent, i, max(1, a.cfg.LEDCount-1))
			frame[i] = mixColors(wheel, tint, 72)
		}
		if err := a.renderFrame(frame); err != nil {
			log.Printf("render error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			shift = (shift + 7) % 256
		}
	}
}

func (a *App) runTheaterChase(ctx context.Context, primary uint32, accent uint32, stepDelay time.Duration) {
	ticker := time.NewTicker(stepDelay)
	defer ticker.Stop()

	offset := 0
	for {
		frame := make([]uint32, a.cfg.LEDCount)
		for i := 0; i < a.cfg.LEDCount; i++ {
			slot := (i + offset) % 4
			switch slot {
			case 0:
				frame[i] = primary
			case 2:
				frame[i] = accent
			default:
				frame[i] = scaleColor(blendColor(primary, accent, 1, 2), 20)
			}
		}
		if err := a.renderFrame(frame); err != nil {
			log.Printf("render error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			offset = (offset + 1) % 3
		}
	}
}

func (a *App) runFade(ctx context.Context, from uint32, to uint32, steps int, stepDelay time.Duration) {
	if steps < 1 {
		steps = 1
	}

	for step := 0; step <= steps; step++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		color := blendColor(from, to, step, steps)
		if err := a.setAll(color); err != nil {
			log.Printf("render error: %v", err)
			return
		}
		time.Sleep(stepDelay)
	}
}

func (a *App) runBlink(ctx context.Context, color uint32, flashes int, stepDelay time.Duration, settle uint32) {
	for i := 0; i < flashes; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := a.setAll(color); err != nil {
			log.Printf("render error: %v", err)
			return
		}
		time.Sleep(stepDelay)

		if err := a.setAll(0x000000); err != nil {
			log.Printf("render error: %v", err)
			return
		}
		time.Sleep(stepDelay)
	}

	if err := a.setAll(settle); err != nil {
		log.Printf("render error: %v", err)
	}
}

func colorForSystem(system string) uint32 {
	return themeForSystem(system).Primary
}

func themeForSystem(system string) SystemTheme {
	key := normalizeSystemKey(system)
	if key == "" {
		return defaultSystemTheme()
	}
	if alias, ok := systemAliases[key]; ok {
		key = alias
	}
	if color, ok := systemBaseColors[key]; ok {
		return newSystemTheme(key, color)
	}
	return defaultSystemTheme()
}

func newSystemTheme(key string, primary uint32) SystemTheme {
	primary = normalizeThemeColor(primary)
	accent := systemAccentColors[key]
	if accent == 0 {
		accent = deriveAccent(primary)
	}
	accent = normalizeThemeColor(accent)
	if accent == primary {
		accent = normalizeThemeColor(deriveAccent(primary))
	}
	return SystemTheme{
		Primary: primary,
		Accent:  accent,
	}
}

func defaultSystemTheme() SystemTheme {
	return SystemTheme{
		Primary: normalizeThemeColor(defaultPrimaryColor),
		Accent:  normalizeThemeColor(defaultAccentColor),
	}
}

func normalizeSystemKey(system string) string {
	var b strings.Builder
	b.Grow(len(system))
	for _, r := range system {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func fallbackColor(system string) uint32 {
	key := normalizeSystemKey(system)
	if key == "" {
		return 0x202020
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	sum := h.Sum32()

	r := uint8(64 + (sum & 0x7f))
	g := uint8(64 + ((sum >> 8) & 0x7f))
	b := uint8(64 + ((sum >> 16) & 0x7f))
	return packColor(r, g, b)
}

func deriveAccent(color uint32) uint32 {
	r, g, b := unpackColor(color)
	accent := packColor(boostChannel(g), boostChannel(b), boostChannel(r))
	if accent == color {
		accent = packColor(boostChannel(255-r), boostChannel(255-g), boostChannel(255-b))
	}
	return accent
}

func scaleColor(color uint32, level uint8) uint32 {
	r, g, b := unpackColor(color)
	return packColor(scaleChannel(r, level), scaleChannel(g, level), scaleChannel(b, level))
}

func normalizeThemeColor(color uint32) uint32 {
	r, g, b := unpackColor(color)
	peak := max(int(r), max(int(g), int(b)))
	if peak == 0 {
		return defaultPrimaryColor
	}
	if peak >= minThemeBrightness {
		return color
	}

	scale := float64(minThemeBrightness) / float64(peak)
	return packColor(
		scaleUpChannel(r, scale),
		scaleUpChannel(g, scale),
		scaleUpChannel(b, scale),
	)
}

func mixColors(from uint32, to uint32, ratio uint8) uint32 {
	fromR, fromG, fromB := unpackColor(from)
	toR, toG, toB := unpackColor(to)
	mix := func(a uint8, b uint8) uint8 {
		return uint8((uint16(a)*(255-uint16(ratio)) + uint16(b)*uint16(ratio)) / 255)
	}
	return packColor(mix(fromR, toR), mix(fromG, toG), mix(fromB, toB))
}

func blendColor(from uint32, to uint32, step int, steps int) uint32 {
	fromR, fromG, fromB := unpackColor(from)
	toR, toG, toB := unpackColor(to)

	blend := func(a uint8, b uint8) uint8 {
		return uint8((int(a)*(steps-step) + int(b)*step) / steps)
	}

	return packColor(blend(fromR, toR), blend(fromG, toG), blend(fromB, toB))
}

func colorWheel(position int) uint32 {
	position %= 256
	if position < 0 {
		position += 256
	}

	switch {
	case position < 85:
		return packColor(uint8(255-position*3), uint8(position*3), 0)
	case position < 170:
		position -= 85
		return packColor(0, uint8(255-position*3), uint8(position*3))
	default:
		position -= 170
		return packColor(uint8(position*3), 0, uint8(255-position*3))
	}
}

func unpackColor(color uint32) (uint8, uint8, uint8) {
	return uint8(color >> 16), uint8(color >> 8), uint8(color)
}

func packColor(r uint8, g uint8, b uint8) uint32 {
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

func scaleChannel(value uint8, level uint8) uint8 {
	return uint8((uint16(value) * uint16(level)) / 255)
}

func boostChannel(value uint8) uint8 {
	boosted := int(value) + 72
	if boosted > 255 {
		return 255
	}
	return uint8(boosted)
}

func scaleUpChannel(value uint8, scale float64) uint8 {
	scaled := int(float64(value) * scale)
	if scaled > 255 {
		return 255
	}
	return uint8(scaled)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func readStateFile(path string) map[string]string {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return out
}
