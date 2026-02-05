package main

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"net"
	"log"
	"strconv"
	"syscall"
)

const (
	timeout = 2 * time.Second
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		port, err := strconv.Atoi(os.Args[2])
		if err != nil {
            log.Fatal("Kunne ikke parse port:", err)
        }
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
		runPrimary(0, addr.Port)
	}
}

func runPrimary(startCounter int, listenPort int) {
	time.Sleep(100 * time.Millisecond)

	backupPort := listenPort + 1
	backupAddr := &net.UDPAddr{
		Port: backupPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	cmd := exec.Command("konsole", "-e", "go", "run", "main.go", "backup", strconv.Itoa(backupPort))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Start()
	time.Sleep(time.Second)

	counter := startCounter
    ticker := time.NewTicker(500 * time.Millisecond)
    heartbeat := time.NewTicker(500 * time.Millisecond)
    
    conn, err := net.DialUDP("udp", nil, backupAddr)
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
			runPrimary(counter, addr.Port)
			return
		}
		fmt.Sscanf(string(buf[:n]), "%d", &counter)
	}
}