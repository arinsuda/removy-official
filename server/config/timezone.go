package config

import (
	"time"
	// Embeds the IANA database so AppLocation resolves on hosts without zoneinfo.
	_ "time/tzdata"
)

// AppTimeZone is the calendar the product runs on: voting round edges and the day a media view counts for.
const AppTimeZone = "Asia/Bangkok"

// AppLocation is loaded once; tzdata is embedded so this cannot fail at runtime.
var AppLocation = mustLoadLocation(AppTimeZone)

// LocalDay returns midnight of t's calendar day in AppLocation.
func LocalDay(t time.Time) time.Time {
	local := t.In(AppLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, AppLocation)
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("config: cannot load time zone " + name + ": " + err.Error())
	}
	return loc
}
