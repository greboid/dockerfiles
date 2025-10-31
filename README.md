# Dockerfiles from the ground up

This started with a fork of [csmith/dockerfiles](https://github.com/csmith/dockerfiles), but eventually
the templating code was moved out of the repo into [contempt](https://github.com/csmith/contempt). It continued using contempt for some time but now moved away from contempt and templated Containerfiles and is now using
[melange](https://github.com/chainguard-dev/melange) to build APK packages, which are then assembled into images with [apko](https://github.com/chainguard-dev/apko).

## What? Why?

This is a collection of containers for various software projects I want to run, built from the ground up.

Most projects have either official docker images or third-party contributions, but they're always a bit
hit-or-miss on how they work, what base images are used, etc. Third-party images often lag behind releases
or just die off without warning, too. Building everything out from scratch ensures the images are standard,
and can follow upstream updates as quickly or slowly as required.

I'm using these production services, but won't vouch for their stability or usability for anyone else's 
purposes. Feel free to use them, and report any issues you do find, but at your own risk!

## Images

All images are available at `reg.g5d.dev/<name>`. Only the latest tag is built.

Each images aims for the following:

**Reproducible** - if the same image is rebuilt at any time on any machine it will produce the same
image. This is nice to have, but it is quite challenging and makes little difference in day-to-day
operations.

**Non-root** - As far as possible these images do not run as root, this reduces the risks.

**Minimal** - the image contains only the bare essentials required. No leftovers, nothing irrelevant,
no bloat. For base images this definition is a bit hazy as they'll contain things that might be used
in downstream images. For applications this generally means they're statically compiled and run in the
"base" image.

This isn't always possible, some images do require root, its not always possible to make them fully
reproducible and some images are quite bloated but do try as closely as possible.
