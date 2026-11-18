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

	startrekdate "github.com/StarFleets/go-startrek-date"
)

func utc(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		panic(err)
	}
	return t
}

// Fixed vectors from the published formula; the fraction is the elapsed hundredths of the day,
// rounded down.
var vectors = []struct {
	at   string
	want string
}{
	{"1998-04-15T00:00:00", "00000.00"},
	{"1998-04-16T00:00:00", "00003.00"},
	{"1998-05-01T06:30:00", "00044.27"},
	{"1999-04-15T00:00:00", "01000.00"},
	{"2000-02-29T12:00:00", "01872.50"},
	{"2000-12-31T23:59:59", "02708.99"},
	{"2001-01-01T00:00:00", "02718.00"},
	{"2024-02-29T00:00:00", "25822.00"},
	{"2024-12-31T12:00:00", "26658.50"},
	{"2025-01-01T00:00:00", "26734.00"},
	{"2026-11-18T12:00:00", "28614.50"},
	{"2026-11-18T19:44:07", "28614.82"},
	{"2026-11-18T23:59:59", "28614.99"},
	{"2100-03-01T00:00:00", "101945.00"},
}

func TestOfficial(t *testing.T) {
	for _, v := range vectors {
		got, err := startrekdate.Official(utc(v.at))
		if err != nil {
			t.Fatalf("%s: unexpected error %v", v.at, err)
		}
		if got.String() != v.want {
			t.Errorf("%s: got %s, want %s", v.at, got, v.want)
		}
	}
}

func TestOfficialUsesUTC(t *testing.T) {
	zone := time.FixedZone("EST", -5*3600)
	local := time.Date(2026, 11, 18, 7, 0, 0, 0, zone) // 12:00 UTC
	got, err := startrekdate.Official(local)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "28614.50" {
		t.Errorf("got %s, want 28614.50", got)
	}
}

func TestOfficialBeforeFoundingDate(t *testing.T) {
	_, err := startrekdate.Official(utc("1998-04-14T23:59:59"))
	if !errors.Is(err, startrekdate.ErrBeforeEpoch) {
		t.Errorf("got %v, want ErrBeforeEpoch", err)
	}
}

func TestDateFields(t *testing.T) {
	got, err := startrekdate.Official(utc("2026-11-18T19:44:07"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Whole != 28614 || got.Hundredths != 82 {
		t.Errorf("got %d.%d, want 28614.82", got.Whole, got.Hundredths)
	}
}

func TestNextTick(t *testing.T) {
	cases := []struct{ at, want string }{
		{"2026-11-18T00:00:00", "2026-11-18T00:14:24"},
		{"2026-11-18T00:14:23", "2026-11-18T00:14:24"},
		{"2026-11-18T00:14:24", "2026-11-18T00:28:48"},
		{"2026-11-18T23:45:35", "2026-11-18T23:45:36"},
		{"2026-11-18T23:50:00", "2026-11-19T00:00:00"},
		{"2026-11-18T23:59:59", "2026-11-19T00:00:00"},
	}
	for _, c := range cases {
		if got := startrekdate.NextTick(utc(c.at)); !got.Equal(utc(c.want)) {
			t.Errorf("NextTick(%s) = %s, want %s", c.at, got.Format("15:04:05"), c.want)
		}
	}
	if startrekdate.TickInterval != 864*time.Second {
		t.Errorf("TickInterval = %s, want 14m24s", startrekdate.TickInterval)
	}
}

type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time { return f.now }

func TestSourceNow(t *testing.T) {
	clock := &fakeClock{now: utc("2026-11-18T12:00:00")}
	var p startrekdate.Provider = startrekdate.New(clock)
	got, err := p.Now()
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "28614.50" {
		t.Errorf("got %s", got)
	}
	clock.now = utc("2026-11-18T18:00:00")
	got, _ = p.Now()
	if got.String() != "28614.75" {
		t.Errorf("got %s after the clock moved", got)
	}
}

func TestCacheServesOneValuePerTick(t *testing.T) {
	clock := &fakeClock{now: utc("2026-11-18T12:00:01")}
	calls := 0
	counting := startrekdate.ProviderFunc(func() (startrekdate.Date, error) {
		calls++
		return startrekdate.Official(clock.now)
	})
	cache := startrekdate.NewCache(counting, clock)

	first, _ := cache.Now()
	clock.now = utc("2026-11-18T12:10:00") // same tick (12:00:00 to 12:14:24)
	second, _ := cache.Now()
	if calls != 1 || first != second {
		t.Fatalf("calls = %d, first %s, second %s: want one computation", calls, first, second)
	}
	clock.now = utc("2026-11-18T12:14:24") // the next tick
	third, _ := cache.Now()
	if calls != 2 || third.String() != "28614.51" {
		t.Errorf("calls = %d, third %s: want a fresh value 28614.51", calls, third)
	}
}

func TestCacheDoesNotKeepErrors(t *testing.T) {
	clock := &fakeClock{now: utc("1998-01-01T00:00:00")}
	cache := startrekdate.NewCache(startrekdate.New(clock), clock)
	if _, err := cache.Now(); !errors.Is(err, startrekdate.ErrBeforeEpoch) {
		t.Fatalf("got %v, want ErrBeforeEpoch", err)
	}
	clock.now = utc("1998-04-15T00:00:00")
	got, err := cache.Now()
	if err != nil || got.String() != "00000.00" {
		t.Errorf("got %s, %v: an error must not be cached", got, err)
	}
}

func TestSystemClock(t *testing.T) {
	before := time.Now()
	got := startrekdate.SystemClock{}.Now()
	if got.Before(before) || time.Since(got) > time.Minute {
		t.Errorf("SystemClock.Now() = %s, want about now", got)
	}
}
