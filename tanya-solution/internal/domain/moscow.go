package domain

import (
	"time"
	_ "time/tzdata" // Europe/Moscow без системной базы часовых поясов
)

var Moscow = mustLoadLocation("Europe/Moscow")

func MoscowDate(t time.Time) string {
	return t.In(Moscow).Format(time.DateOnly)
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
