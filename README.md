# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)

Service that can download video files, transforms them in to audio files and then converts them to RSS audio podcast feeds.

Currently the service only supports YouTube videos.

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)

### Optional

- [Python](https://www.python.org/) (used only as starting script)

Use `make` to run the project. Make is typically installed out of the box on Linux and Mac.

If you do not have it and run on Windows, you can directly install it from [gnuwin32](https://gnuwin32.sourceforge.net/packages/make.htm) or via `winget`

```PowerShell
winget install GnuWin32.Make
```

If you want to run the project without Docker, you can install [Golang](https://go.dev/doc/install)

## How to Use

### Start only this Service

You can use `make` to start the service

```bash
make start
```

Or just use native docker commands. E.g.

```bash
docker build . -t v2p
docker run --rm -p 8080:8080 v2p
```

### Resources

All downloaded resources will be places in directories `resources`.
Podcasts will be structured in directories which have the name of the channel the video belongs to.

### Start with EMail Webhook

This option allow to start an additional service that continuously pulls an email address and used the mail link in the content of unread mails as input for the service.
The configuration of this service is described [here](https://github.com/jo-hoe/go-mail-webhook-service).

## Example Requests

Below are a few examples of requests to the service.

### Add Items

These APIs can be used to add videos to the service.
Note that the service excepts individual video links, as well as a playlists.

#### Add Single Item

```bash
curl -H "Content-Type: application/json" --data '{"url":"https://www.youtube.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-"}' http://localhost:8080/v1/addItem
```

#### Add Multiple Items

```bash
curl -H "Content-Type: application/json" --data '{"urls":["https://www.youtube.com/watch?v=BRnwg3dpboc", "https://www.youtube.com/watch?v=_fWrJ4WHz_g"]}' http://localhost:8080/v1/addItems
```

### Get All Feeds

Use this API to get a list of all feeds.

```bash
curl -H "Content-Type: application/json" http://localhost:8080/v1/feeds
```

## Linting

Project used golangci-lint for linting.

### Installation

See <https://golangci-lint.run/usage/install/>

### Execution

Run the linting locally by executing

```bash
golangci-lint run ./...
```

in the working directory

## Relevant Links

- [ID3 Tags](https://www.exiftool.org/TagNames/ID3.html)
- [example podcast](https://feeds.libsyn.com/230510/rss)
