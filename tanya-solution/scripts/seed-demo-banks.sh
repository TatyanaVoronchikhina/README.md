#!/usr/bin/env sh
# Синтетические снимки demo-bank-a/b (USD/RUB cash Moscow) с текущим временем — для convert в data-demo.
set -eu

DATA_DIR="${DATA_DIR:-./data-demo}"
case "$(basename "$DATA_DIR")" in
  data-demo) ;;
  *) echo "DATA_DIR must point to data-demo, got $DATA_DIR" >&2; exit 1 ;;
esac

day="$(TZ=Europe/Moscow date +%Y-%m-%d)"
observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
file_ts="$(date -u +%Y%m%dT%H%M%S).000000000Z"
dir="$DATA_DIR/archive/$day"
mkdir -p "$dir"

write_bank() { # source bank buy sell
  tmp="$dir/.$1.tmp"
  {
    echo "observed_at,source_at,source,kind,bank,city,channel,base,quote,buy_rate,sell_rate,reference_rate,source_url"
    echo "$observed_at,$observed_at,$1,bank,$2,Moscow,cash,USD,RUB,$3,$4,,https://example.org/$1"
  } > "$tmp"
  mv "$tmp" "$dir/$1-$file_ts.csv"
  echo "$dir/$1-$file_ts.csv"
}

write_bank demo-bank-a "Demo Bank A" 80 90
write_bank demo-bank-b "Demo Bank B" 81 89
