#!/bin/bash
set -e

trap_with_arg() {
    func="$1" ; shift
    for sig ; do
        trap "$func $sig" "$sig"
    done
}

func_trap() {
    echo "Trapped: $1"
}

trap_with_arg func_trap USR1 USR2

echo "Send signals to PID $$ and type [enter] when done."
sleep 20
# read # Wait so the script doesn't exit.
