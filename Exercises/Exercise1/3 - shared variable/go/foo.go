// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
)

type msg int

const (
    inc msg = iota
    dec
    get
)

var req = make(chan msg)
var resp = make(chan int)
var done = make(chan struct{})

func numberServer() {
    i := 0
    for m := range req {
        switch m {
        case inc:
            i++
        case dec:
            i--
        case get:
            resp<- i
        }
    }
}

func incrementing() {
    //TODO: increment i 1000000 times
    for j := 0; j < 1000000; j++ {
        req<- inc
    }
    done<- struct{}{} //Uses zero memory
}

func decrementing() {
    //TODO: decrement i 1000000 times
    for j := 0; j < 1000000; j++ {
        req<- dec
    }
    done<- struct{}{}
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(16)

    // TODO: Spawn both functions as goroutines
    go numberServer()
    go incrementing()
    go decrementing()

    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    <-done
    <-done
    req<- get
    Println("The magic number is:", <-resp)
}
