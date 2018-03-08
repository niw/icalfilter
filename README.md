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

`icalfilterd` is a HTTP deamon that filters a large iClaendar file on the fly.

`make docker` creates a `icalfilter` docker image.

`Dockerfile` works on [Heroku](https://www.heroku.com) [Container Registry & Runtime](https://devcenter.heroku.com/articles/container-registry-and-runtime).
