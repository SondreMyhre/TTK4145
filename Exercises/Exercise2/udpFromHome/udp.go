package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	announcePort = 30000
	echoRecvPort = 20000 // server_wfh lytter her
	echoSendPort = 20001 // server_wfh svarer hit
)

func getServerIP() net.IP {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: announcePort})
	if err != nil {
		log.Fatal("Failed to bind UDP announce socket:", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	buf := make([]byte, 1024)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			// Timeout -> fallback (vanlig hjemme)
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return nil
			}
			continue
		}

		msg := strings.TrimSpace(string(buf[:n]))
		fmt.Println("ANNOUNCE:", msg)
		return from.IP
	}
}

func UDPReceiver(port int) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		log.Fatal("ListenUDP receiver failed:", err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		fmt.Printf("ECHO recv from %v: %q\n", from, string(buf[:n]))
	}
}

func UDPSender(serverIP net.IP, port int) {
	raddr := &net.UDPAddr{IP: serverIP, Port: port}
	conn, err := net.DialUDP("udp4", nil, raddr)
	if err != nil {
		log.Fatal("DialUDP failed:", err)
	}
	defer conn.Close()

	for i := 0; i < 3 ; i++ {
		msg := fmt.Sprintf("hei %d", i)
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println("UDP write error:", err)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	serverIP := getServerIP()
	if serverIP == nil {
		// WFH: broadcast kan feile → server kjører lokalt
		serverIP = net.ParseIP("127.0.0.1")
		fmt.Println("No announce received; fallback to", serverIP)
	} else {
		fmt.Println("Server IP:", serverIP)
	}

	// Viktig: lytte på reply-port før sending
	go UDPReceiver(echoSendPort)
	time.Sleep(100 * time.Millisecond)

	// Send til serverens recv-port
	go UDPSender(serverIP, echoRecvPort)

	select {}
}
