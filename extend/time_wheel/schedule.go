package cherryTimeWheel

import (
	"time"
)

// Scheduler determines the execution plan of a task.
type Scheduler interface {
	// Next returns the next execution time after the given (previous) time.
	// It will return a zero time if no next time is scheduled.
	//
	// All times must be UTC.
	Next(time.Time) time.Time
}

type EverySchedule struct {
	Interval time.Duration
}

func (s *EverySchedule) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}

type FixedDateSchedule struct {
	Hour, Minute, Second int
}

func (s *FixedDateSchedule) Next(prev time.Time) time.Time {
	hour := prev.Hour()
	if s.Hour >= 0 {
		hour = s.Hour
	}

	fixedTime := time.Date(
		prev.Year(),
		prev.Month(),
		prev.Day(),
		hour,
		s.Minute,
		s.Second,
		0,
		prev.Location(),
	)

	remain := fixedTime.UnixNano() - prev.UnixNano()
	if remain > 0 {
		return prev.Add(time.Duration(remain))
	}

	if s.Hour >= 0 {
		return fixedTime.Add(24 * time.Hour)
	}

	return fixedTime.Add(time.Hour)
}
