package time

import "time"

func NewTicker(duration time.Duration, immediately bool) *CustomTicker {

	ticker := &CustomTicker{}

	ch := make(chan time.Time)
	ticker.C = ch

	go func() {
		stdTicker := time.NewTicker(duration)
		for t := range stdTicker.C {
			ch <- t
		}
	}()

	if immediately {
		go func() {
			ch <- time.Now()
		}()
	}

	return ticker
}

type CustomTicker struct {
	C <-chan time.Time
}
