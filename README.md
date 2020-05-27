icalfilter
==========

A simple tool that filters a large iCalendar file by removing past events.

Build
-----

Check out repository and update `libical` submodule.

    git submodule update --init --recursive

Use `make` to build a `icalfiler` CLI command.

    make

HTTP server
-----------

`icalfilterd` is a HTTP daemon that filters a large iCalendar file on the fly.

`make docker` creates a `icalfilter` docker image.

The docker image works on [Heroku](https://www.heroku.com) [Container Registry & Runtime](https://devcenter.heroku.com/articles/container-registry-and-runtime) or [Google App Engine](https://cloud.google.com/appengine/) [Flexible Environment](https://cloud.google.com/appengine/docs/flexible/custom-runtimes/).
