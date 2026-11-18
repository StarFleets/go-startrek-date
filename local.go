package startrekdate

/*
Apache License 2.0

Copyright 2026 Shane Froebel & the rabbit hole tech, llc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	// ErrNoZone is returned when a time zone is needed and none was given.
	ErrNoZone = errors.New("startrekdate: no time zone given")
	// ErrBadDate is returned for a date that is not a real calendar date written NNNN-NN-NN.
	ErrBadDate = errors.New("startrekdate: date must be a real date written YYYY-MM-DD")
	// ErrBadTime is returned for a time that is not HH:MM or HH:MM:SS on a 24-hour clock.
	ErrBadTime = errors.New("startrekdate: time must be written HH:MM or HH:MM:SS on a 24-hour clock")
	// ErrUnknownZone is returned for a zone name the time zone database does not know.
	ErrUnknownZone = errors.New("startrekdate: unknown time zone")
	// ErrNonexistentTime is returned for a local time that never happened because the clocks
	// skipped it when daylight saving time began.
	ErrNonexistentTime = errors.New("startrekdate: that local time does not exist in the zone")
)

// Local returns the stardate for the calendar day and wall-clock time t has in loc, instead of
// in UTC. The days since the founding date are counted on the zone's own calendar, the divisor is
// 365 or 366 by the zone's year, and the fraction is the wall-clock reading (a day in which the
// clocks change is still read as 24 hours). Local(t, time.UTC) is the official stardate.
//
// A day before the founding date, on the zone's calendar, returns ErrBeforeEpoch. A nil loc
// returns ErrNoZone.
func Local(t time.Time, loc *time.Location) (Date, error) {
	if loc == nil {
		return Date{}, ErrNoZone
	}
	t = t.In(loc)
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	if day.Before(epoch) {
		return Date{}, ErrBeforeEpoch
	}
	days := int(day.Sub(epoch) / (24 * time.Hour))
	yearLength := 365.0
	if isLeap(t.Year()) {
		yearLength = 366.0
	}
	whole := int(math.Round(1000 / yearLength * float64(days)))
	seconds := t.Hour()*3600 + t.Minute()*60 + t.Second()
	return Date{Whole: whole, Hundredths: seconds * 100 / 86400}, nil
}

// Reading is one moment shown two ways.
type Reading struct {
	// Time is the moment.
	Time time.Time
	// Zone is the zone the Local value was worked out in.
	Zone *time.Location
	// Official is the official stardate: the formula counts UTC days and the UTC time of day,
	// so it is the same for everyone at the same moment.
	Official Date
	// Local is the stardate for the calendar day and wall-clock time in Zone, which differs
	// from Official for part of every day. It is a convenience, not the official value.
	Local Date
}

// String formats both values with their labels, for example
// "28614.82 official (UTC), 28614.61 local (America/New_York)".
func (r Reading) String() string {
	return fmt.Sprintf("%s official (UTC), %s local (%s)", r.Official, r.Local, r.Zone)
}

// Convert returns t as a Reading for loc. Both values are needed, so a moment whose UTC day or
// whose day in loc is before the founding date returns ErrBeforeEpoch. A nil loc returns
// ErrNoZone.
func Convert(t time.Time, loc *time.Location) (Reading, error) {
	official, err := Official(t)
	if err != nil {
		return Reading{}, err
	}
	local, err := Local(t, loc)
	if err != nil {
		return Reading{}, err
	}
	return Reading{Time: t, Zone: loc, Official: official, Local: local}, nil
}

// ParseLocal turns a typed local date, time and zone into a moment in that zone. date is
// YYYY-MM-DD. clock is HH:MM or HH:MM:SS on a 24-hour clock, and empty means midnight. zone is
// an IANA name such as "America/New_York", or "UTC".
//
// A local time that never happened (the hour skipped when daylight saving time begins) returns
// ErrNonexistentTime. A local time that happened twice (the hour repeated when it ends) is read
// as its first occurrence, the daylight-saving one.
func ParseLocal(date, clock, zone string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, ErrBadDate
	}
	h, m, s, err := parseClock(clock)
	if err != nil {
		return time.Time{}, err
	}
	if zone == "" {
		return time.Time{}, fmt.Errorf("%w: empty name", ErrUnknownZone)
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrUnknownZone, zone)
	}
	return resolve(d.Year(), d.Month(), d.Day(), h, m, s, loc)
}

func parseClock(clock string) (h, m, s int, err error) {
	if clock == "" {
		return 0, 0, 0, nil
	}
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, perr := time.Parse(layout, clock); perr == nil {
			return t.Hour(), t.Minute(), t.Second(), nil
		}
	}
	return 0, 0, 0, ErrBadTime
}

// resolve finds the instant whose wall clock in loc reads the given time. It tries the offsets
// in force a day before and a day after, which covers every daylight-saving change, and keeps
// the earliest instant that reads back the same.
func resolve(year int, month time.Month, day, h, m, s int, loc *time.Location) (time.Time, error) {
	wall := time.Date(year, month, day, h, m, s, 0, time.UTC)
	var best time.Time
	for _, probe := range []time.Duration{-24 * time.Hour, 24 * time.Hour} {
		_, offset := wall.Add(probe).In(loc).Zone()
		candidate := wall.Add(-time.Duration(offset) * time.Second).In(loc)
		if candidate.Year() != year || candidate.Month() != month || candidate.Day() != day ||
			candidate.Hour() != h || candidate.Minute() != m || candidate.Second() != s {
			continue
		}
		if best.IsZero() || candidate.Before(best) {
			best = candidate
		}
	}
	if best.IsZero() {
		return time.Time{}, ErrNonexistentTime
	}
	return best, nil
}
