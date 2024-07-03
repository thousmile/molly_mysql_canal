package main

import (
	"fmt"
	"testing"
	"time"
)

func Test001(t *testing.T) {
	s1 := "last_update_user"
	fmt.Println(lowerCamelCase(s1))
	fmt.Println(upperCamelCase(s1))
}

func Test002(t *testing.T) {
	newVal, _ := time.Parse(time.DateOnly, "2024-07-03")
	if newVal.Hour() == 0 && newVal.Minute() == 0 && newVal.Second() == 0 {
		fmt.Println(newVal.Format(time.DateOnly))
	} else if newVal.Year() == 0 && newVal.Month() == 0 && newVal.Day() == 0 {
		fmt.Println(newVal.Format(time.TimeOnly))
	} else {
		fmt.Println(newVal.Format(time.RFC3339))
	}
}
