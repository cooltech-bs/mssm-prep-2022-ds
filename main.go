package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
)

func main() {
	// Use bufio here because it is ~20x faster than unbuffered reading on my
	// test environment (WSL-Arch).
	covered, err := doFireOneLifeguard(bufio.NewReader(os.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println(covered)
}

const (
	GuardUnknown = iota
	GuardStart
	GuardEnd
)

type Event struct {
	Time    int
	GuardID int
	Type    int // whether the guard starts or ends their shift
}

type Events []Event

func (e Events) Len() int {
	return len(e)
}

func (e Events) Less(i, j int) bool {
	return e[i].Time < e[j].Time
}

func (e Events) Swap(i, j int) {
	// TODO: optimize constant factor: avoid moving whole structs
	e[i], e[j] = e[j], e[i]
}

func doFireOneLifeguard(fin io.Reader) (covered int, err error) {
	var (
		numberOfGuards, i int
		start, end        int
		shifts            Events
		currentOnDuty     map[int]struct{}
		aloneOnDutyTime   map[int]int
		tmpGuardID        int
		minTAlone         = math.MaxInt
	)
	_, err = fmt.Fscanf(fin, "%d\n", &numberOfGuards)
	if err != nil {
		return
	}
	shifts = make(Events, 2*numberOfGuards)

	// Read shift data and convert to events
	for i = 0; i < numberOfGuards; i++ {
		_, err = fmt.Fscanf(fin, "%d %d\n", &start, &end)
		if err != nil {
			return
		}
		shifts[i*2] = Event{
			Time:    start,
			GuardID: i,
			Type:    GuardStart,
		}
		shifts[i*2+1] = Event{
			Time:    end,
			GuardID: i,
			Type:    GuardEnd,
		}
	}
	sort.Stable(shifts) // O(nlogn)

	currentOnDuty = make(map[int]struct{})
	aloneOnDutyTime = make(map[int]int)
	for i = 0; i < len(shifts); i++ { // O(n)
		if len(currentOnDuty) == 0 {
			// Must be a start-shift event
			start = shifts[i].Time
		} else if len(currentOnDuty) == 1 {
			// Get the only guard on duty
			for tmpGuardID = range currentOnDuty {
				break
			}
			// Sum up the time intervals when he/she is alone on duty
			aloneOnDutyTime[tmpGuardID] += shifts[i].Time - shifts[i-1].Time
		}
		switch shifts[i].Type {
		case GuardStart:
			currentOnDuty[shifts[i].GuardID] = struct{}{}
		case GuardEnd:
			delete(currentOnDuty, shifts[i].GuardID)
		}
		if len(currentOnDuty) == 0 {
			end = shifts[i].Time
			covered += end - start
		}
	}

	if len(aloneOnDutyTime) == 0 {
		return
	}
	// O(m) (m is the average number of guards on duty at any time)
	// According to birthday paradox, m should be proportional to sqrt(n)
	for _, tAlone := range aloneOnDutyTime {
		if minTAlone > tAlone {
			minTAlone = tAlone
		}
	}
	covered -= minTAlone
	return
}
