package main

import (
	localsingle "TTK4145/LocalSingleElevator"
)

func main() {
	go localsingle.Run()

	select {}
}
