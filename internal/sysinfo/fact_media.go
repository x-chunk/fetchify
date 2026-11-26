package sysinfo

import (
	"context"
	"strings"
)

// FactMedia is what a media player is playing.
const FactMedia Fact = "media"

// Media is the track of a media player.
type Media struct {
	// Artist is the artist of the track, or "".
	Artist string
	// Title is the title of the track.
	Title string
	// Status is "Playing", "Paused" or "Stopped".
	Status string
}

// mediaFormat asks playerctl for the fields of Media, separated by tabs,
// which do not occur in metadata.
const mediaFormat = "{{status}}\t{{artist}}\t{{title}}"

func init() {
	collectors[FactMedia] = fact(collectMedia, func(i *Info) *Value[Media] { return &i.Media })
}

// collectMedia asks playerctl for the current track. The fact is Unknown
// when playerctl is not installed and Absent when no player runs or the
// player has no track.
func collectMedia(ctx context.Context, s *session) Value[Media] {
	out, err := s.run(ctx, "playerctl", "metadata", "--format", mediaFormat)
	if notInstalled(err) {
		return Value[Media]{}
	}
	if err != nil {
		return absent[Media]()
	}
	line, _, _ := strings.Cut(string(out), "\n")
	fields := strings.Split(line, "\t")
	if len(fields) != 3 || strings.TrimSpace(fields[2]) == "" {
		return absent[Media]()
	}
	return present(Media{
		Status: strings.TrimSpace(fields[0]),
		Artist: strings.TrimSpace(fields[1]),
		Title:  strings.TrimSpace(fields[2]),
	})
}
