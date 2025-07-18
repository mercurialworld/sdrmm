package utils

import (
	"log"
)

func PanicOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
