# Dockerfiles from the ground up

This started with a fork of [csmith/dockerfiles](https://github.com/csmith/dockerfiles), but eventually 
the templating code was moved out of the repo into [contempt](https://github.com/csmith/contempt).

This is now just a collection of dockerfiles I use in my infrastructure.

## What? Why?

This is a collection of Dockerfiles for various software projects I want to run, built from the ground up.

Most projects have either official docker images or third-party contributions, but they're always a bit
hit-or-miss on how they work, what base images are used, etc. Third-party images often lag behind releases
or just die off without warning, too. Building everything out from scratch ensures the images are standard,
and can follow upstream updates as quickly or slowly as required.

I'm using these production services, but won't vouch for their stability or usability for anyone else's 
purposes. Feel free to use them, and report any issues you do find, but at your own risk!

## Images

All images are available at `reg.g5d.dev/<name>`. Only the latest tag is built.

Each images aims for the following:

**Reproducible** - if the same Dockerfile is rebuilt at any time on any machine it will produce the same
image. This is nice to have, but it is quite challenging and makes little difference in day-to-day
operations. (It also requires the use of a tool like `buildah` that can set layer timestamps; `docker build`
cannot make reproducible images.)

**Non-root** - the entrypoint for the image is invoked as a non-root user. Base images are marked as N/A.
Other images probably drop root later either via a script or as part of the process itself, but it's
preferable for it to happen in the image.

**Minimal** - the image contains only the bare essentials required. No leftovers, nothing irrelevant,
no bloat. For base images this definition is a bit hazy as they'll contain things that might be used
in downstream images. For applications this generally means they're statically compiled and run in the
"base" image.

This isn't always possible, some images do require root, its not always possible to make them fully
reproducible and some images are quite bloated but do try as closely as possible.
