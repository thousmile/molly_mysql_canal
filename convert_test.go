package main

import (
	"fmt"
	"testing"
)

func Test001(t *testing.T) {
	s1 := "last_update_user"
	fmt.Println(LowerCamelCase(s1))
	fmt.Println(UpperCamelCase(s1))
}
