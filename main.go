package main

import "github.com/darkonie/wikiracer/supervisor"

func main() {
	err := supervisor.Start()
	if err != nil {
		panic(err)
	}
}
