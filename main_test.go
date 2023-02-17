package main

import (
	"time"
)

func Example_printStats() {
	stats := StatsMap{
		time.Date(2022, time.December, 14, 0, 0, 0, 0, time.UTC): map[int64]struct{}{1: sentinel, 2: sentinel, 3: sentinel},
		time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC):   map[int64]struct{}{1: sentinel, 2: sentinel, 3: sentinel},
		// 4 unique devs in November
		time.Date(2022, time.November, 1, 0, 0, 0, 0, time.UTC):  map[int64]struct{}{1: sentinel, 2: sentinel, 3: sentinel},
		time.Date(2022, time.November, 21, 0, 0, 0, 0, time.UTC): map[int64]struct{}{2: sentinel, 3: sentinel, 4: sentinel},
	}
	printStats(stats)
	// Output:
	// 2022-11, 4
	// 2022-12, 3
	// 2023-1, 3
}
