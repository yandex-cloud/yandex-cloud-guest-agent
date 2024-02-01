// Please note that the code below is modified by YANDEX LLC

package main

import "time"

type Clock interface {
	Now() time.Time
}
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

var (
	clockInstance = &realClock{}
)
