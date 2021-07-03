package helper

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

var SqlDateTimeFormat = "2006-01-02 15:04:05"
var SqlDateFormat = "2006-01-02"

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

func GetAuth(req *http.Request) (username string, password string, status int) {
	auth := strings.SplitN(req.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		status = 401
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		status = 401
		return
	}

	return pair[0], pair[1], 200
}

func ContainsInt(list []int, value int) bool {
	for _, entry := range list {
		if entry == value {
			return true
		}
	}
	return false
}
