#!/usr/bin/env bash
set -euo pipefail

HOST="${HOST:-http://localhost:8080}"
ENDPOINT="${ENDPOINT:-/place_bid/}"
USER_PREFIX="${USER_PREFIX:-user}"
COUNT="${COUNT:-5}"
BASE_BID="${BASE_BID:-200}"
BASE_BW="${BASE_BW:-100}"
STEP_BID="${STEP_BID:-20}"
STEP_BW="${STEP_BW:-10}"

for ((i=1; i<=COUNT; i++)); do
  user="${USER_PREFIX}${i}"
  bid=$((BASE_BID + i * STEP_BID))
  bw=$((BASE_BW + i * STEP_BW))
  url="${HOST}${ENDPOINT}?user=${user}&bidValue=${bid}&bandwidthValue=${bw}"
  echo "Sending: ${url}"
  curl -s "$url"
  echo ""
  sleep 0.2
 done
