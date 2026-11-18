# go-startrek-date

> 🧭 Star Date calculation for Star Trek themed sites: the Starfleets official stardate as a small Go package

## ✨ Highlights

- 🖖 **The official Starfleets stardate:** the one starfleets.net has shown since 1998, ported from its old PHP and checked against fixed vectors.
- 🧪 **Testable:** the time comes from a `Clock` you inject, and `Provider` is a small interface to fake.
- 🌍 **Any date, any zone:** `Convert` shows the official value and the value for a time zone, side by side, and `ParseLocal` turns a typed date, time and zone into a moment.
- ⏱️ **Tick helper:** `NextTick` and `Cache` serve one value per hundredth of a day (14 minutes 24 seconds).
- 📦 **No dependencies:** the standard library only.

## 📦 Install

```bash
go get github.com/StarFleets/go-startrek-date
```

## 🚀 Use

```go
package main

import (
	"fmt"

	startrekdate "github.com/StarFleets/go-startrek-date"
)

func main() {
	clock := startrekdate.SystemClock{}
	source := startrekdate.New(clock)
	cache := startrekdate.NewCache(source, clock) // one computation per tick

	date, err := cache.Now()
	if err != nil {
		panic(err)
	}
	fmt.Println(date) // for example 28614.82
}
```

`Official(t)` is the pure function behind it. A date prints as `NNNNN.NN`.

### Any date and time, and a time zone

```go
at, err := startrekdate.ParseLocal("2026-11-18", "14:44:07", "America/New_York")
if err != nil {
	panic(err)
}
reading, err := startrekdate.Convert(at, at.Location())
if err != nil {
	panic(err)
}
fmt.Println(reading) // 28614.82 official (UTC), 28614.61 local (America/New_York)
```

The **official** stardate counts UTC days and the UTC time of day, so it is the same for everyone at the same moment. The **local** value (`Local`) uses the calendar day and wall-clock time in the zone, with the 365 or 366 divisor taken from the local year; it is a convenience and can differ from the official value for part of every day. `Reading` always labels both.

Edge cases: a moment whose UTC day or zone day is before 15 April 1998 returns `ErrBeforeEpoch`; a local time skipped by a daylight saving change returns `ErrNonexistentTime`; a local time that happens twice is read as its first occurrence (the daylight saving one). Zone names are IANA names such as `America/New_York`, or `UTC`.

## 🧮 The formula

- The whole part is the number of days since the founding date, 15 April 1998, times 1000, divided by 365 (366 when the date's year is a leap year), rounded to the nearest whole number. The divisor follows the year of the date, so the number steps at New Year, as it always did on the site.
- The fraction is the elapsed hundredths of the UTC day, 0 to 99.
- `Official` reads the time in UTC; `Local` reads it in the zone you give.

Two deliberate differences from the old PHP: the fraction is the elapsed hundredths (rounded down), so it never prints `.100` as the old rounding did after 23:52:48, and a time before the founding date returns `ErrBeforeEpoch` instead of a malformed number. There is no "unofficial Star Trek" date.

## 🛠 Develop

```bash
task build    # go build ./...
task test     # go test ./...
task lint     # gofmt check + golangci-lint + yamllint
task license  # check Apache-2.0 headers (golic)
```

## ⚖️ License

Apache-2.0 (c) 2026 Shane Froebel & the rabbit hole tech, llc
