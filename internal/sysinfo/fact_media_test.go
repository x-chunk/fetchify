package sysinfo

import (
	"context"
	"errors"
	"testing"
)

func TestMedia(t *testing.T) {
	cmd := "playerctl metadata --format " + mediaFormat
	tests := []struct {
		name string
		src  Source
		want Value[Media]
	}{
		{"playing", machine{commands: map[string]string{cmd: "Playing\tDaft Punk\tAround the World\n"}}.source(),
			present(Media{Artist: "Daft Punk", Title: "Around the World", Status: "Playing"})},
		{"no artist", machine{commands: map[string]string{cmd: "Paused\t\tPodcast 12\n"}}.source(),
			present(Media{Title: "Podcast 12", Status: "Paused"})},
		{"no title", machine{commands: map[string]string{cmd: "Stopped\t\t\n"}}.source(), absent[Media]()},
		{"no player", Source{Run: func(context.Context, string, ...string) ([]byte, error) {
			return nil, errors.New("exit status 1")
		}}, absent[Media]()},
		{"not installed", machine{}.source(), Value[Media]{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Collect(context.Background(), tt.src, []Fact{FactMedia}, Options{}).Media; got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
