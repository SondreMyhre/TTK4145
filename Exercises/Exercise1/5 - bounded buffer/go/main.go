
package main

import "fmt"
import "time"


func producer(buf chan int, done chan int){

    for i := 0; i < 10; i++ {
        time.Sleep(100 * time.Millisecond)
        fmt.Printf("[producer]: pushing %d\n", i)
        // TODO: push real value to buffer
        buf<- i
    }

    close(buf)
    done<- 1
}

func consumer(buf chan int, done chan int){

    time.Sleep(1 * time.Second)
    for v:= range buf {
        // i := 0 //TODO: get real value from buffer
        // i = <-buf
        fmt.Printf("[consumer]: %d\n", v)
        time.Sleep(50 * time.Millisecond)
    }

    done<- 1
    
}


func main(){
    
    // TODO: make a bounded buffer
    var buf = make(chan int, 5)
    var done = make(chan int)
    go consumer(buf, done)
    go producer(buf, done)
    
    // select {}
    <-done
    <-done
}