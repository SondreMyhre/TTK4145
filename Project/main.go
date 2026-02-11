package main

import (
	localsingle "Project/LocalSingleElevator"
)

func main() {
	go localsingle.Run()

	select {}
}
