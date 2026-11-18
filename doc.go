// Package startrekdate computes the Starfleets official stardate.
//
// The whole part counts days since the founding date, 15 April 1998, scaled so a year is about
// 1000 units: days * 1000 / 365 (366 in a leap year, judged on the year of the date asked
// about), rounded to the nearest whole number. The fraction is the elapsed hundredths of the
// UTC day. A date prints as NNNNN.NN, for example 28614.82.
//
// The official value counts UTC days and the UTC time of day. Local gives the same formula for the
// calendar day and wall-clock time in a time zone, Convert returns both side by side as a
// Reading, and ParseLocal turns a typed date, time and zone into a moment.
//
// Official is a pure function. Source reads the time from a Clock you inject, which keeps
// callers testable. Cache serves one value per hundredth of a day (TickInterval, 14 minutes
// 24 seconds) so a busy server computes it once per tick.
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
