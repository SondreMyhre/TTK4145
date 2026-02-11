package main

import(
	"fmt"
	"time"
	"log"
	"net"
	"os/exec"
	"strconv"
)

const(
	localhost = "localhost:50000"
	timeout = 2 //i sekunder
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", localhost)
	if err != nil {
		log.Fatalln(err)
	}

	counterStr := runBackup(addr)
	startNewBackupProg()
	runPrimary(counterStr)
}

func startNewBackupProg() {
	cmd := exec.Command("cmd", "/C", "start", "cmd", "/K", "go run main.go")
	err := cmd.Start()
	if err != nil {
		log.Fatalln(err)
	}
}

func runPrimary(counterStr string) {
	conn, err := net.Dial("udp4", localhost)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	counter, err := strconv.Atoi(counterStr)
	if err != nil {
		log.Fatalln("Error converting string to int:", err)
	}
	for i:= counter; ;i++ {
		fmt.Println(i)
		time.Sleep(500 * time.Millisecond)

		message := strconv.Itoa(i)
		_, err = conn.Write([]byte(message))
		if err != nil {
			log.Fatalln("Error sending message:", err)
		}
	}
}

func runBackup(lnAddr *net.UDPAddr) string {
	buf := make([]byte, 1024)
	conn, err := net.ListenUDP("udp4", lnAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	lastMsg := "0"
	for {
		conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			ne, ok := err.(net.Error)
			if ok && ne.Timeout() {
				return lastMsg
			}
			log.Fatalln("Error reading from server:", err)
		}
		lastMsg = string(buf[:n])
	}
}
