package main

import (
	"testfsnotify/modules/lg"
	"testfsnotify/modules/tm"
)

func main() {
	lg.NewLogger("test", "app.log")
	lg.Info("test initiate")

	tm.Testif()
	tm.Tester()
}
