package startrekdate_test

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
	"testing"
	"time"
	_ "time/tzdata"

	startrekdate "github.com/StarFleets/go-startrek-date"
)

func zone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

// Vectors come from the old PHP formula applied to the zone's own calendar day and wall clock.
var zoneVectors = []struct {
	utc, zone, official, local string
}{
	{"2026-11-18T19:44:07", "America/New_York", "28614.82", "28614.61"},
	{"2026-11-19T02:00:00", "America/New_York", "28616.08", "28614.87"},
	{"2026-11-18T20:00:00", "Asia/Kolkata", "28614.83", "28616.06"},
	{"2024-12-31T23:00:00", "Pacific/Auckland", "26658.95", "26734.50"},
	{"2000-12-31T23:00:00", "Pacific/Kiritimati", "02708.95", "02718.54"},
	{"2026-11-18T12:00:00", "UTC", "28614.50", "28614.50"},
	{"2026-03-08T07:30:00", "America/New_York", "27915.31", "27915.14"},
	{"2026-11-01T05:30:00", "America/New_York", "28567.22", "28567.06"},
}

func TestLocal(t *testing.T) {
	for _, v := range zoneVectors {
		got, err := startrekdate.Local(utc(v.utc), zone(t, v.zone))
		if err != nil {
			t.Fatalf("%s %s: %v", v.utc, v.zone, err)
		}
		if got.String() != v.local {
			t.Errorf("%s in %s: got %s, want %s", v.utc, v.zone, got, v.local)
		}
	}
}

func TestLocalNeedsAZone(t *testing.T) {
	if _, err := startrekdate.Local(utc("2026-11-18T12:00:00"), nil); !errors.Is(err, startrekdate.ErrNoZone) {
		t.Errorf("got %v, want ErrNoZone", err)
	}
}

func TestLocalBeforeFoundingDate(t *testing.T) {
	// 1998-04-15 03:00 UTC is still 14 April in Los Angeles.
	_, err := startrekdate.Local(utc("1998-04-15T03:00:00"), zone(t, "America/Los_Angeles"))
	if !errors.Is(err, startrekdate.ErrBeforeEpoch) {
		t.Errorf("got %v, want ErrBeforeEpoch", err)
	}
	// 1998-04-14 23:00 UTC is already 15 April in Tokyo.
	got, err := startrekdate.Local(utc("1998-04-14T23:00:00"), zone(t, "Asia/Tokyo"))
	if err != nil || got.String() != "00000.33" {
		t.Errorf("got %s, %v, want 00000.33", got, err)
	}
}

func TestConvert(t *testing.T) {
	for _, v := range zoneVectors {
		loc := zone(t, v.zone)
		r, err := startrekdate.Convert(utc(v.utc), loc)
		if err != nil {
			t.Fatalf("%s %s: %v", v.utc, v.zone, err)
		}
		if r.Official.String() != v.official || r.Local.String() != v.local {
			t.Errorf("%s in %s: got official %s local %s, want %s and %s", v.utc, v.zone, r.Official, r.Local, v.official, v.local)
		}
		if r.Zone != loc || !r.Time.Equal(utc(v.utc)) {
			t.Errorf("%s in %s: reading kept the wrong zone or time", v.utc, v.zone)
		}
	}
}

func TestConvertString(t *testing.T) {
	r, err := startrekdate.Convert(utc("2026-11-18T19:44:07"), zone(t, "America/New_York"))
	if err != nil {
		t.Fatal(err)
	}
	want := "28614.82 official (UTC), 28614.61 local (America/New_York)"
	if r.String() != want {
		t.Errorf("got %q, want %q", r.String(), want)
	}
}

func TestConvertErrors(t *testing.T) {
	if _, err := startrekdate.Convert(utc("2026-11-18T12:00:00"), nil); !errors.Is(err, startrekdate.ErrNoZone) {
		t.Errorf("nil zone: got %v, want ErrNoZone", err)
	}
	// Official side before the epoch (UTC day is 14 April) while the Tokyo side is 15 April.
	if _, err := startrekdate.Convert(utc("1998-04-14T23:00:00"), zone(t, "Asia/Tokyo")); !errors.Is(err, startrekdate.ErrBeforeEpoch) {
		t.Errorf("official before epoch: got %v, want ErrBeforeEpoch", err)
	}
	// Local side before the epoch while the official side is 15 April.
	if _, err := startrekdate.Convert(utc("1998-04-15T03:00:00"), zone(t, "America/Los_Angeles")); !errors.Is(err, startrekdate.ErrBeforeEpoch) {
		t.Errorf("local before epoch: got %v, want ErrBeforeEpoch", err)
	}
}

func TestParseLocal(t *testing.T) {
	cases := []struct {
		date, clock, zone, want string // want is the UTC instant
	}{
		{"2026-11-18", "14:44:07", "America/New_York", "2026-11-18T19:44:07"},
		{"2026-11-18", "14:44", "America/New_York", "2026-11-18T19:44:00"},
		{"2026-11-18", "", "America/New_York", "2026-11-18T05:00:00"},
		{"2026-11-18", "01:30", "Asia/Kolkata", "2026-11-17T20:00:00"},
		{"2026-11-18", "12:00", "UTC", "2026-11-18T12:00:00"},
		{"2024-02-29", "12:00", "UTC", "2024-02-29T12:00:00"},
		// Ambiguous time in the autumn overlap: the first occurrence (daylight time) is used.
		{"2026-11-01", "01:30", "America/New_York", "2026-11-01T05:30:00"},
		// A time just after the overlap is unambiguous.
		{"2026-11-01", "02:30", "America/New_York", "2026-11-01T07:30:00"},
		// A half-hour zone overlap: Lord Howe ends daylight time on 5 April 2026 at 02:00 local, back to 01:30.
		{"2026-04-05", "01:45", "Australia/Lord_Howe", "2026-04-04T14:45:00"},
	}
	for _, c := range cases {
		got, err := startrekdate.ParseLocal(c.date, c.clock, c.zone)
		if err != nil {
			t.Errorf("%v: %v", c, err)
			continue
		}
		if !got.Equal(utc(c.want)) {
			t.Errorf("%v: got %s, want %s", c, got.UTC().Format("2006-01-02T15:04:05"), c.want)
		}
		if got.Location().String() != c.zone {
			t.Errorf("%v: result zone %s, want %s", c, got.Location(), c.zone)
		}
	}
}

func TestParseLocalErrors(t *testing.T) {
	cases := []struct {
		name, date, clock, zone string
		want                    error
	}{
		{"bad date", "18-11-2026", "12:00", "UTC", startrekdate.ErrBadDate},
		{"impossible date", "2026-02-30", "12:00", "UTC", startrekdate.ErrBadDate},
		{"empty date", "", "12:00", "UTC", startrekdate.ErrBadDate},
		{"bad time", "2026-11-18", "25:00", "UTC", startrekdate.ErrBadTime},
		{"text time", "2026-11-18", "noon", "UTC", startrekdate.ErrBadTime},
		{"unknown zone", "2026-11-18", "12:00", "Mars/Olympus_Mons", startrekdate.ErrUnknownZone},
		{"empty zone", "2026-11-18", "12:00", "", startrekdate.ErrUnknownZone},
		{"spring gap", "2026-03-08", "02:30", "America/New_York", startrekdate.ErrNonexistentTime},
		{"half-hour gap", "2026-10-04", "02:15", "Australia/Lord_Howe", startrekdate.ErrNonexistentTime},
	}
	for _, c := range cases {
		if _, err := startrekdate.ParseLocal(c.date, c.clock, c.zone); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, err, c.want)
		}
	}
}

func TestParseLocalThenConvert(t *testing.T) {
	at, err := startrekdate.ParseLocal("2026-11-18", "14:44:07", "America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	r, err := startrekdate.Convert(at, at.Location())
	if err != nil {
		t.Fatal(err)
	}
	if r.Official.String() != "28614.82" || r.Local.String() != "28614.61" {
		t.Errorf("got %s", r)
	}
}
