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
	"sync"
	"time"
)

// TickInterval is one hundredth of a day: the time between changes of Date.Hundredths.
const TickInterval = 864 * time.Second

// ErrBeforeEpoch is returned for a time before the founding date, which has no stardate.
var ErrBeforeEpoch = errors.New("startrekdate: time is before the founding date 1998-04-15")

var epoch = time.Date(1998, time.April, 15, 0, 0, 0, 0, time.UTC)

// Date is a Starfleets official stardate.
type Date struct {
	// Whole is the integer part.
	Whole int
	// Hundredths is the elapsed hundredths of the UTC day, 0 to 99.
	Hundredths int
}

// String formats the date as NNNNN.NN, the whole part padded with zeros to five digits.
func (d Date) String() string {
	return fmt.Sprintf("%05d.%02d", d.Whole, d.Hundredths)
}

// Official returns the official stardate for t, which is read in UTC. A time before the
// founding date returns ErrBeforeEpoch.
func Official(t time.Time) (Date, error) {
	return Local(t, time.UTC)
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// NextTick returns the first moment after t at which Date.Hundredths changes.
func NextTick(t time.Time) time.Time {
	t = t.UTC()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	into := t.Sub(midnight)
	return midnight.Add((into/TickInterval + 1) * TickInterval)
}

// Clock tells the time. The zero value of a caller's dependency should be a real clock; tests
// pass a fake.
type Clock interface {
	Now() time.Time
}

// Provider returns the current stardate.
type Provider interface {
	Now() (Date, error)
}

// ProviderFunc adapts a function to Provider.
type ProviderFunc func() (Date, error)

// Now calls f.
func (f ProviderFunc) Now() (Date, error) { return f() }

// Source is a Provider that reads the time from a Clock.
type Source struct {
	clock Clock
}

// New returns a Source reading clock.
func New(clock Clock) *Source {
	return &Source{clock: clock}
}

// Now returns the official stardate for the clock's current time.
func (s *Source) Now() (Date, error) {
	return Official(s.clock.Now())
}

// Cache serves one Provider value until the next tick. It is safe for concurrent use. Errors
// are never cached.
type Cache struct {
	provider Provider
	clock    Clock

	mu      sync.Mutex
	value   Date
	expires time.Time
}

// NewCache returns a Cache over provider, using clock to decide when a value has expired.
func NewCache(provider Provider, clock Clock) *Cache {
	return &Cache{provider: provider, clock: clock}
}

// Now returns the cached value while the clock is before the next tick, and asks the provider
// again after it.
func (c *Cache) Now() (Date, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock.Now()
	if now.Before(c.expires) {
		return c.value, nil
	}
	value, err := c.provider.Now()
	if err != nil {
		return Date{}, err
	}
	c.value, c.expires = value, NextTick(now)
	return value, nil
}

// SystemClock is the Clock that reads the real time.
type SystemClock struct{}

// Now returns the current time.
func (SystemClock) Now() time.Time { return time.Now() }
