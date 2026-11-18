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
	"fmt"
	"time"

	startrekdate "github.com/StarFleets/go-startrek-date"
)

func ExampleOfficial() {
	date, _ := startrekdate.Official(time.Date(2026, time.November, 18, 19, 44, 7, 0, time.UTC))
	fmt.Println(date)
	// Output: 28614.82
}

func ExampleOfficial_beforeFounding() {
	_, err := startrekdate.Official(time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC))
	fmt.Println(err)
	// Output: startrekdate: time is before the founding date 1998-04-15
}

func ExampleNextTick() {
	at := time.Date(2026, time.November, 18, 12, 0, 1, 0, time.UTC)
	fmt.Println(startrekdate.NextTick(at).Format("15:04:05"))
	// Output: 12:14:24
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func ExampleNewCache() {
	clock := fixedClock{now: time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)}
	cache := startrekdate.NewCache(startrekdate.New(clock), clock)
	date, _ := cache.Now()
	fmt.Println(date)
	// Output: 28614.50
}

func ExampleConvert() {
	loc, _ := time.LoadLocation("America/New_York")
	at := time.Date(2026, time.November, 18, 19, 44, 7, 0, time.UTC)
	reading, _ := startrekdate.Convert(at, loc)
	fmt.Println(reading)
	// Output: 28614.82 official (UTC), 28614.61 local (America/New_York)
}

func ExampleParseLocal() {
	at, _ := startrekdate.ParseLocal("2026-11-18", "14:44:07", "America/New_York")
	fmt.Println(at.UTC().Format(time.RFC3339))
	// Output: 2026-11-18T19:44:07Z
}

func ExampleParseLocal_skippedHour() {
	_, err := startrekdate.ParseLocal("2026-03-08", "02:30", "America/New_York")
	fmt.Println(err)
	// Output: startrekdate: that local time does not exist in the zone
}
