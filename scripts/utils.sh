#!/usr/bin/env bash

function is_val_number() {
    local res="$1"
    local num="$2"
    re='^[0-9]+$'
    if ! [[ ${num} =~ $re ]] ; then
        eval ${res}="'0'"
    else
        eval ${res}="'1'"
    fi
}
