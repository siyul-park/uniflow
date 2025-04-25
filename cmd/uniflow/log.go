package main

import "log"

func must[T any](val T, err error) T {
	if err != nil {
		fatal(err)
	}
	return val
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
