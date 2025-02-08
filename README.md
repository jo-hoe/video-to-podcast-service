# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/video-to-podcast-service)](https://goreportcard.com/report/github.com/jo-hoe/video-to-podcast-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/video-to-podcast-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/video-to-podcast-service?branch=main)

A service that can download video files, transform them into audio files, and then convert them to RSS audio podcast feeds.

Currently, the service only supports YouTube videos.

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)

### Optional

- [Python](https://www.python.org/) (used only as starting script)

Run the project using `make`. Make is typically installed by default on Linux and Mac.

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

All downloaded resources will be placed in directories `resources`.
Podcasts will be structured in directories with the name of the channel the video belongs to.

## Example Requests

Below are a few examples of requests to the service.

### Add Items

These APIs can be used to add videos to the service.
Note that the service accepts individual video links and a playlist.

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

### Start with EMail Webhook

This option allows you to start an additional service that continuously pulls an email address and uses the mail link in the content of unread mails as input for the service.
The configuration of this service is described [here](https://github.com/jo-hoe/video-to-podcast-service/blob/main/mail-webhook-config/config.yaml).

## Linting

The project uses golangci-lint for linting.

### Installation

See <https://golangci-lint.run/usage/install/>

### Execution

Run the linting locally by executing

```bash
golangci-lint run ./...
```

in the working directory.

## Limitations

Google blocks certain IPs.
This includes IPs form hyperscalers (such as AWS).
Trying to download youtube videos from such a IP results in an error such as `403`.
The lib used in this project returns `Error:can't bypass age restriction: login required to confirm your age`.
You can find more details regarding this issue [here](https://github.com/kkdai/youtube/issues/343#issuecomment-2347950479).

## Future Work

- one can implement itunes tags to get a pic for each podcast element (however, the lib does not support this, implementation requires generating the xml and not using lib)
  - example `<itunes:image href="http://....png"/>`
- create `index.html` with all podcasts and qr codes
- provide ticketing return/progression via api

## Relevant Links

- [ID3 Tags](https://www.exiftool.org/TagNames/ID3.html)
- [example podcast](https://feeds.libsyn.com/230510/rss)
