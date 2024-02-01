// Please note that the code below is modified by YANDEX LLC

package main

import "time"

type mockClock struct {
	t time.Time
}

func newMockClock(t time.Time) mockClock {
	return mockClock{t}
}

func (m mockClock) Now() time.Time {
	return m.t
}
