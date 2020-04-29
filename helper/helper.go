package helper

import "time"

var SqlDateFormat = "2006-01-02 15:04:05"

func Schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func (e *LoadError) Error() string {
	return e.Msg
}

type LoadError struct {
	Msg string
}
