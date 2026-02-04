package main

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"net"
	"log"
	"strconv"
)

const (
	timeout = 2 * time.Second
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		port, _ := strconv.Atoi(os.Args[2])
		addr := &net.UDPAddr{
			Port: port,
			IP:   net.ParseIP("127.0.0.1"),
		}
		runBackup(addr)
	} else {
		addr := &net.UDPAddr{
			Port: 20002,
			IP:   net.ParseIP("127.0.0.1"),
		}
		runPrimary(0, addr)
	}
}

func runPrimary(startCounter int, addr *net.UDPAddr) {
	time.Sleep(100 * time.Millisecond)

	nextAddr := &net.UDPAddr{
		Port: addr.Port + 1,
		IP:   addr.IP,
	}

	cmd := exec.Command("konsole", "-e", "go", "run", "main.go", "backup", nextAddr.String())
	cmd.Start()
	time.Sleep(time.Second)

	counter := startCounter
    ticker := time.NewTicker(500 * time.Millisecond)
    heartbeat := time.NewTicker(500 * time.Millisecond)
    
    conn, err := net.DialUDP("udp", nil, nextAddr)
	if err != nil {
        log.Fatal("Primary kunne ikke åpne UDP:", err)
    }
    defer conn.Close()
    
    for {
        select {
        case <-ticker.C:
            counter++
            fmt.Println(counter)
            
        case <-heartbeat.C:
            _, err := conn.Write([]byte(strconv.Itoa(counter)))
            if err != nil {
                log.Fatal("Send feil:", err)
            }
        }
    }


}

func runBackup(addr *net.UDPAddr) {
	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
        log.Fatal("Backup kunne ikke lytte på port:", err)
    }
	defer conn.Close()
	
	fmt.Printf("Backup startet og lytter på %s\n", addr.String())
	counter := 0
	buf := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Primary døde! Blir primary...")
			conn.Close()
			time.Sleep(500 * time.Millisecond)
			runPrimary(counter, addr)
			return
		}
		fmt.Sscanf(string(buf[:n]), "%d", &counter)
	}
}