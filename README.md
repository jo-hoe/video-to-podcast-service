# go-audio-rss-feeder

```bash
ffprobe -i .\video.mp4 -v quiet -of json -show_entries format
```

example podcast: <https://feeds.libsyn.com/230510/rss>
ID3: <https://www.exiftool.org/TagNames/ID3.html>

## Example Requests

### Add Playlist

```bash
curl -H "Content-Type: application/json" --data '{"url":"https://www.youtube.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-"}' http://localhost:8080/v1/addItem
```

### Get All Feeds

```bash
curl -H "Content-Type: application/json" http://localhost:8080/v1/feeds
```
