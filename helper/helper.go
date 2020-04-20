package helper

import "time"

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
