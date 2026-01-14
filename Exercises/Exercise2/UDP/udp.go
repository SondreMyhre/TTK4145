package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func getServerIP() net.IP {

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 30000})
	if err != nil {
		fmt.Println("Failed to bind UDP socket:", err)
		return nil
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := strings.TrimSpace(string(buf[:n]))

		fmt.Println(msg)
		time.Sleep(500 * time.Millisecond)
		return from.IP
	}

}

func UDPReceiver(port int) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		log.Fatalf("ListenUDP failed: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, from, err := conn.ReadFromUDP(buf)
        if err != nil { continue }
        fmt.Printf("UDP receive from %v: %q\n", from, string(buf[:n]))
	}
}

func UDPSender(serverIP net.IP, port int) {
    raddr:= &net.UDPAddr{IP: serverIP, Port: port}
    conn, err := net.DialUDP("udp4", nil, raddr)
    if err != nil { log.Fatal(err) }
    defer conn.Close()

    for i := 0; ; i++ {
        msg := fmt.Sprintf("hei %v", i)
        _, _ = conn.Write([]byte(msg))
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
	serverIP := getServerIP()
	fmt.Println("Server IP:", serverIP)

    n := 13
    port := 20000 + n

    go UDPReceiver(port)
    time.Sleep(200 * time.Millisecond)
    go UDPSender(serverIP, port)
    fmt.Println("Listening")
    select{}

}
