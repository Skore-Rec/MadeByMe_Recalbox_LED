#!/bin/bash
BASE=/recalbox/share/userscripts/recalbox-ledd
LOG=/recalbox/share/system/logs/recalbox-ledd.log
BIN="$BASE/recalbox-ledd"
TMP_BIN="/tmp/recalbox-ledd"

if [ ! -f "$BIN" ]; then
  echo "binary not found: $BIN" >> "$LOG"
  exit 1
fi

exec >> "$LOG" 2>&1

if ! cp "$BIN" "$TMP_BIN"; then
  echo "[$(date)] failed to copy binary to $TMP_BIN"
  exit 1
fi

if ! chmod +x "$TMP_BIN"; then
  echo "[$(date)] failed to chmod +x $TMP_BIN"
  exit 1
fi

echo "[$(date)] starting recalbox-ledd"
exec "$TMP_BIN" -mqtt tcp://127.0.0.1:1883 -topic Recalbox/EmulationStation/Event -state /tmp/es_state.inf -count 16 -gpio 18 -brightness 96
