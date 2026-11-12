#!/bin/sh
# Mark root=gpt-auto as handled; gptroot-mount.sh mounts it.
case "$(getarg root=)" in
    gpt-auto)
        rootok=1
        ;;
esac
