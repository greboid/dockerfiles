#!/bin/sh
# root=gpt-auto is resolvable (the mount-stage hook mounts the
# discoverable root partition); mark it handled so the init does not
# die at the cmdline stage.
case "$(getarg root=)" in
    gpt-auto)
        rootok=1
        ;;
esac
