// Use `go run foo.go` to run your program

//go:build ignore
// +build ignore

package main

import (
    . "fmt"
    "runtime"
    // "time"
    "sync"
)

var i = 0
var mutex = sync.Mutex{}
var wg = sync.WaitGroup{}
var intChannel1 = make(chan int)
var intChannel2 = make(chan int)

func incrementing() {
    //TODO: increment i 1000000 times
    for j := 0; j < 1000000; j++ {
        mutex.Lock()
        i++
        mutex.Unlock()
    }
    // wg.Done()
    intChannel1<-1
}

func decrementing() {
    //TODO: decrement i 1000000 times
    for j := 0; j < 999999; j++ {
        mutex.Lock()
        i--
        mutex.Unlock()
    }
    // wg.Done()
    intChannel2<-1
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(2)
	
    // TODO: Spawn both functions as goroutines
    // wg.Add(2)

    go incrementing()
    go decrementing()

    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    // time.Sleep(500*time.Millisecond)
    // wg.Wait()
    
    <-intChannel1
    <-intChannel2

    Println("The magic number is:", i)
}
