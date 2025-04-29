package utils

import "log"

func Go(f func()) {
	go func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
		f()
	}()
}
