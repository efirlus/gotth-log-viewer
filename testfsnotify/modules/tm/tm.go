package tm

import (
	"fmt"
	"testfsnotify/modules/lg"
	"time"
)

func Testif() {
	t := time.Now()
	s := fmt.Sprintf("current time is %v", t)
	lg.Info(s)
}

func Tester() {
	t := time.Now()
	s := fmt.Errorf("error occurs on %s", t)
	lg.Err("error", s)
}
