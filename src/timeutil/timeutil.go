package timeutil

import (
	"fmt"
	"time"

	"github.com/quasilyte/ge/langs"
)

func FormatDateISO8601(d time.Time, withTime bool) string {
	s := fmt.Sprintf("%04d-%02d-%02d", d.Year(), d.Month(), d.Day())
	if withTime {
		s += fmt.Sprintf(" %02d:%02d", d.Hour(), d.Minute())
	}
	return s
}

func FormatDurationCompact(d time.Duration) string {
	d = d.Round(time.Second)
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func FormatDuration(dict *langs.Dictionary, d time.Duration) string {
	d = d.Round(time.Second)
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	if hours >= 1 {
		return fmt.Sprintf("%d%s %d%s %d%s",
			hours, dict.Get("game.value.hour"), minutes, dict.Get("game.value.minute"), seconds, dict.Get("game.value.second"))
	}
	if minutes >= 1 {
		return fmt.Sprintf("%d%s %d%s", minutes, dict.Get("game.value.minute"), seconds, dict.Get("game.value.second"))

	}
	return fmt.Sprintf("%d%s", seconds, dict.Get("game.value.second"))
}
