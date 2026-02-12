// Package logging for reusable logging
package logging

import "fmt"

func Log(prefix string, message string) {
	fmt.Println(prefix + " " + message)
}
