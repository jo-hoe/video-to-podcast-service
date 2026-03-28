package twitch

import (
	"testing"
)

func TestTwitchAudioDownloader_IsVideoSupported(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		d    *TwitchAudioDownloader
		args args
		want bool
	}{
		{
			name: "vod link",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://www.twitch.tv/videos/2345678901"},
			want: true,
		},
		{
			name: "vod link without www",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://twitch.tv/videos/2345678901"},
			want: true,
		},
		{
			name: "clip via channel path",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://www.twitch.tv/somechannel/clip/SomeClipSlug-AbCdEf"},
			want: true,
		},
		{
			name: "clip via clips subdomain",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://clips.twitch.tv/SomeClipSlug-AbCdEf"},
			want: true,
		},
		{
			name: "channel page is not supported",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://www.twitch.tv/somechannel"},
			want: false,
		},
		{
			name: "youtube link is not supported",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://www.youtube.com/watch?v=jNQXAC9IVRw"},
			want: false,
		},
		{
			name: "non-existing link is not supported",
			d:    &TwitchAudioDownloader{},
			args: args{url: "https://not-existing.com/videos/123"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.IsVideoSupported(tt.args.url); got != tt.want {
				t.Errorf("TwitchAudioDownloader.IsVideoSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
