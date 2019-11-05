#!/bin/sh

if test "$#" -lt 2; then
    return 2
fi

if [ "$1" == "buildrun" ]; then
    make || exit 1;
fi

shift

./pkg $*
