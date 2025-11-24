#!/bin/sh
if [ "$1" = "--gtk3" ]; then
    ./marmalade-gtk3
else
    ./marmalade-gtk4
fi
