package main

//TODO: Get localsingle.Run() to be Go version of elevalgo

import (
	"TTK4145/LocalSingleElevator"
)

func main() {
	go localsingle.Run()

    select {}
}
