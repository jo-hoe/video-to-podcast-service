# Audio RSS Feeder

[![Test Status](https://github.com/jo-hoe/go-audio-rss-feeder/workflows/test/badge.svg)](https://github.com/jo-hoe/go-audio-rss-feeder/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/go-audio-rss-feeder/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-audio-rss-feeder/actions?workflow=lint)

Converts video files to RSS audio podcast feeds.

## Example Requests

### Add Items

#### Add Single Item

```bash
curl -H "Content-Type: application/json" --data '{"url":"https://www.youtube.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-"}' http://localhost:8080/v1/addItem
```

#### Add Multiple Items

```bash
curl -H "Content-Type: application/json" --data '{"urls":["https://www.youtube.com/watch?v=BRnwg3dpboc", "https://www.youtube.com/watch?v=_fWrJ4WHz_g"]}' http://localhost:8080/v1/addItems
```

### Get All Feeds

```bash
curl -H "Content-Type: application/json" http://localhost:8080/v1/feeds
```

## Relevant Links

- [ID3 Tags](https://www.exiftool.org/TagNames/ID3.html)
- [example podcast](https://feeds.libsyn.com/230510/rss)
