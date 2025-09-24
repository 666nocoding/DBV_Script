package main

import (
	"DBV_Script/dbv"
	"log/slog"
)

func main() {
	if err := dbv.Parser(); err != nil {
		slog.Error(err.Error())
		return
	}
	dbv.Run()
}
