#!/bin/sh
set -eu

usage() {
    printf '%s\n' \
        'Usage: sh scripts/build-half-cpu.sh [docker build arguments]' \
        '       sh scripts/build-half-cpu.sh --run [foreground docker run arguments]' \
        '       sh scripts/build-half-cpu.sh --print-budget'
}

mode=build
case "${1:-}" in
    --help|-h) usage; exit 0 ;;
    --run) mode=run; shift ;;
    --print-budget) mode=budget; shift ;;
    '') usage >&2; exit 64 ;;
esac

for argument do
    case "$argument" in
        --cpus|--cpus=*|--cpu-quota|--cpu-quota=*|--cpu-period|--cpu-period=*)
            printf '%s\n' 'CPU limit overrides are not allowed.' >&2
            exit 64 ;;
    esac
    if [ "$mode" = run ]; then
        case "$argument" in
            --detach|--detach=*|-d|-d?*|-i*d*|-t*d*)
                printf '%s\n' 'Detached jobs cannot share the foreground CPU budget lock.' >&2
                exit 64 ;;
        esac
    fi
done

command -v docker >/dev/null
command -v flock >/dev/null
host_cpus=$(getconf _NPROCESSORS_ONLN)
daemon_cpus=$(docker info --format '{{.NCPU}}')
case "$host_cpus:$daemon_cpus" in
    *[!0-9:]*|:*|*:) printf '%s\n' 'Cannot determine the CPU budget.' >&2; exit 1 ;;
esac
[ "$host_cpus" -gt 0 ] && [ "$daemon_cpus" -gt 0 ]
[ "$daemon_cpus" -ge "$host_cpus" ] || host_cpus=$daemon_cpus
if command -v nproc >/dev/null; then
    allowed_cpus=$(nproc)
    [ "$allowed_cpus" -ge "$host_cpus" ] || host_cpus=$allowed_cpus
fi
period=100000
quota=$((host_cpus * 50000))

# A restricted caller must not accidentally grant half of a larger host.
if [ -r /proc/self/cgroup ]; then
    group=$(awk -F: '$1 == "0" { print $3; exit }' /proc/self/cgroup)
    if [ -n "$group" ]; then
        group=/sys/fs/cgroup$group
        while :; do
            if [ -r "$group/cpu.max" ]; then
                read -r available available_period < "$group/cpu.max"
                if [ "$available" != max ]; then
                    group_quota=$(awk -v q="$available" -v p="$available_period" 'BEGIN { printf "%.0f", int(q * 50000 / p) }')
                    [ "$group_quota" -ge "$quota" ] || quota=$group_quota
                fi
            fi
            [ "$group" != /sys/fs/cgroup ] || break
            group=${group%/*}
        done
    fi
fi
if [ "$quota" -lt 1000 ]; then
    printf '%s\n' 'Half of the available CPU capacity is below Docker minimum quota.' >&2
    exit 1
fi
cpus=$(awk -v q="$quota" 'BEGIN { printf "%.6f", q / 100000 }')
if [ "$mode" = budget ]; then
    [ "$#" -eq 0 ] || { usage >&2; exit 64; }
    printf '%s\n' "$cpus"
    exit 0
fi
[ "$#" -gt 0 ] || { usage >&2; exit 64; }

# Hold the same host-wide lock for the entire build or foreground test job.
exec 9>/var/lock/h5gacha-cpu-budget.lock
flock 9
printf 'CPU ceiling: %s cores (quota %s / period %s); other managed jobs are serialized.\n' "$cpus" "$quota" "$period" >&2
if [ "$mode" = run ]; then
    exec docker run --cpus="$cpus" "$@"
fi
exec env DOCKER_BUILDKIT=0 docker build --cpu-period="$period" --cpu-quota="$quota" "$@"
