package analytics

import "time"

type ClickEvent struct {
	Code      string
	IP        string
	UserAgent string
	Referer   string
	ClickedAt time.Time
}
