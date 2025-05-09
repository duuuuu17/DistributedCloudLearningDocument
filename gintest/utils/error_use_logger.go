package my_utils

import (
	"log"
)

func ErrorInLog(err error) {
	if err != nil {
		log.Fatalln("error: ", err)
	}
}
