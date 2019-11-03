#!/bin/bash
log_file=ispend.log

crr_dir="$(dirname "$0")"
source ${crr_dir}/utils.sh

HELP_STR="usage: $0 [-h] display help [-p] port num"
optspec=":ph-:"
while getopts "$optspec" optchar; do
    case "${optchar}" in
        p)
                val="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                port_num="${val}"
                ;;
        h)
            echo "${HELP_STR}" >&2
            exit 2
            ;;
        *)
            if [[ "$OPTERR" != 1 ]] || [[ "${optspec:0:1}" = ":" ]]; then
                echo "Error parsing short flag: '-${OPTARG}'" >&2
                exit 1
            fi

            ;;
    esac
done

if [[ -z "$1" ]]; then
  echo "${HELP_STR}" >&2
  exit 2
fi

if [[ -z "${port_num}" ]]; then
  echo "port num not specified!"
  exit 3
fi

### check port is a valid number
is_val_number port_ok ${port_num}
if ! [[ ${port_ok} =~ 1 ]] ; then
    echo "wrong/corrupt port number: ${port_num}"
    exit 1
fi
###

function rebirth() {
    local port="$1"
    local log_file="$2"

    echo "Killing the server [localhost:${port}]..."
    curl localhost:${port}/harakiri
    echo
    sleep 1s
    echo "Getting born again..."
    echo "run cmd/main.go -port=${port} -logfile=${log_file} &"
    go run cmd/main.go -port=${port} -logfile=${logFile} &
    echo "Server got reborn"
}

rebirth ${port_num} ${log_file}
