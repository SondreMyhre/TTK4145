package main

import (
	localsingle2 "TTK4145/LocalSingleElevator2"
)

func main() {
	go localsingle2.Run()

	select {}
}
