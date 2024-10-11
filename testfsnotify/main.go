package main

import (
	"fmt"
	"reflect"
	"testfsnotify/modules/lg"
	"testfsnotify/modules/tm"
)

func main() {
	lg.NewLogger("test", "app.log")
	lg.Info("test initiate")

	tm.Testif()
	tm.Tester()
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
