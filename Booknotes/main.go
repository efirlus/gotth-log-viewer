package main

import (
	"booknotes/modules/cv"
	"booknotes/modules/fsn"
	"booknotes/modules/lg"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// TODO: lg 모듈 각 모듈에 삽입하기
//
//	stacktrace 되니까 다 log하지 말고 받아넣기
//
// 앞으로도 존나 복사해다가 쓸 fsnotify watcher
// 주요 기능
//  1. 강제 종료도 로깅 가능
//  2. 그럼에도 불구하고 systemctl service 최적화
//  3. 당연하지만 concurrency
//  4. 모든 핸들링을 단 하나의 함수에 뭉쳐놓아 간결함
func main() {

	lg.NewLogger("BookNotes", cv.LogPath)

	err := fsn.InitWatcher()
	if err != nil {
		lg.Fatal("Cannot Initiate Watcher", err)
	}
	defer fsn.CloseWatcher()
	lg.Info("------Program Initiated------")

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	sigmessage := fmt.Sprintf("received signal %v, shutting down...", sig)
	lg.Info(sigmessage)

	done := make(chan bool)
	go func() {
		fsn.CloseWatcher()
		lg.Panic("cleanup completed", nil)
		done <- true
	}()

	select {
	case <-done:
		lg.Info("shutdown completed")
	case <-time.After(30 * time.Second):
		lg.Info("shutdown timed out, force kill")
	}

	os.Exit(0)
}

func printStruct(s interface{}) {
	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		fmt.Printf("%s: %v\n", field.Name, value.Interface())
	}
}
