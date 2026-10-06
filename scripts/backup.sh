#!/bin/sh
set -eu

DIR=/backups
PREFIX=habit_tracker # no hyphen: prune() splits file names on "-"
KEEP_NIGHTLY=14
KEEP_MONTHLY=12
MAX_AGE_MINUTES=1560 # 26 hours

log() { echo "$(date '+%F %T') backup: $*"; }

fresh_dump_exists() {
    find "$DIR" -maxdepth 1 -name "$PREFIX-*.dump" -mmin "-$MAX_AGE_MINUTES" | grep -q .
}

# Names sort chronologically: habit_tracker-2026-10-02_0300.dump
prune() {
    all=$(find "$DIR" -maxdepth 1 -name "$PREFIX-*.dump" | sort)
    nightly=$(echo "$all" | tail -n "$KEEP_NIGHTLY")
    # The first dump of each month, newest twelve months.
    monthly=$(echo "$all" | awk -F- '!seen[$(NF-2) "-" $(NF-1)]++' | tail -n "$KEEP_MONTHLY")
    keep=$(printf '%s\n%s\n' "$nightly" "$monthly")
    echo "$all" | while read -r file; do
        [ -n "$file" ] || continue
        if ! echo "$keep" | grep -qxF "$file"; then
            rm -f "$file"
            log "pruned $(basename "$file")"
        fi
    done
}

run() {
    name="$PREFIX-$(date +%Y-%m-%d_%H%M).dump"
    # A hidden temporary name, renamed only on success, so a half-written
    # file never matches the glob that pruning and the healthcheck rely on.
    partial="$DIR/.$name.partial"
    if ! pg_dump --format=custom --file="$partial"; then
        rm -f "$partial"
        log "FAILED: pg_dump exited non-zero; no backup written"
        exit 1
    fi
    mv "$partial" "$DIR/$name"
    log "wrote $name ($(du -k "$DIR/$name" | cut -f1) KB)"
    prune
}

schedule() {
    # busybox crond does not pass the container environment to its jobs.
    export -p | grep -E '^export (PG[A-Z]+|TZ)=' > /etc/backup.env
    echo "0 3 * * * . /etc/backup.env && sh /scripts/backup.sh run > /proc/1/fd/1 2>&1" > /etc/crontabs/root
    if fresh_dump_exists; then
        log "newest dump is recent; next at 03:00"
    else
        log "no dump in the last 26 hours; taking one now"
        run || true # stay up; the healthcheck reports the failure
    fi
    exec crond -f -l 8
}

case "${1:-}" in
    schedule) schedule ;;
    run) run ;;
    healthcheck) fresh_dump_exists ;;
    *) echo "usage: backup.sh schedule|run|healthcheck" >&2; exit 2 ;;
esac
