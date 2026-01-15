package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	udpDiscoveryPort = 30000
	tcpZeroPort      = 33546
	tcpFixedPort     = 34933
)

// For fixed-size TCP: always send exactly 1024 bytes
func pad1024(msg string) []byte {
	b := make([]byte, 1024)
	copy(b, []byte(msg+"\x00"))
	return b
}

// ---- UDP DISCOVERY ----
// Listens on UDP :30000 and returns the server's IP.
// Tries to parse IP from payload; if not possible, falls back to sender IP.
func discoverServerIP() net.IP {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: udpDiscoveryPort})
	if err != nil {
		log.Fatal("Failed to bind UDP discovery socket:", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		payload := strings.TrimSpace(string(buf[:n]))

		// If server sends its IP as text in payload
		if ip := net.ParseIP(payload); ip != nil {
			fmt.Println("Discovered server IP from payload:", ip)
			return ip
		}

		// Fallback: use sender address
		if from != nil && from.IP != nil {
			fmt.Println("Discovered server IP from sender:", from.IP)
			return from.IP
		}
	}
}

// ---- TCP \0 TERMINATED (port 33546) ----
func runTCPZero(serverIP net.IP) {
	addr := &net.TCPAddr{IP: serverIP, Port: tcpZeroPort}
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		log.Println("TCP(\\0) dial failed:", err)
		return
	}
	fmt.Println("TCP(\\0) connected to", addr.String())

	// Receiver goroutine: read until \0
	go func() {
		defer conn.Close()
		r := bufio.NewReader(conn)
		for {
			msg, err := r.ReadBytes(0)
			if err != nil {
				log.Println("TCP(\\0) read error:", err)
				return
			}
			fmt.Println("TCP(\\0) recv:", string(bytes.TrimRight(msg, "\x00")))
		}
	}()

	// Sender loop (can be its own goroutine too; keeping it here is fine)
	go func() {
		i := 0
		for {
			text := fmt.Sprintf("hello tcp0 %d\x00", i) // MUST end with \0
			_, err := conn.Write([]byte(text))
			if err != nil {
				log.Println("TCP(\\0) write error:", err)
				return
			}
			i++
			time.Sleep(1 * time.Second)
		}
	}()
}

// ---- TCP FIXED 1024 (port 34933) ----
func runTCPFixed(serverIP net.IP) {
	addr := &net.TCPAddr{IP: serverIP, Port: tcpFixedPort}
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		log.Println("TCP(1024) dial failed:", err)
		return
	}
	fmt.Println("TCP(1024) connected to", addr.String())

	// Receiver goroutine: always read exactly 1024 bytes
	go func() {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			_, err := io.ReadFull(conn, buf)
			if err != nil {
				log.Println("TCP(1024) read error:", err)
				return
			}
			fmt.Println("TCP(1024) recv:", string(bytes.TrimRight(buf, "\x00")))
		}
	}()

	// Sender goroutine: always write exactly 1024 bytes
	go func() {
		i := 0
		for {
			payload := pad1024(fmt.Sprintf("hello fixed %d", i))
			_, err := conn.Write(payload)
			if err != nil {
				log.Println("TCP(1024) write error:", err)
				return
			}
			i++
			time.Sleep(1 * time.Second)
		}
	}()
}

// ---- CONNECT BACK ----
// 1) Find local IP used to reach server (no hardcoding)
// 2) Listen on localIP:0 (OS picks a port)
// 3) Tell server: "Connect to: ip:port\0" on TCP \0 port
// 4) Accept incoming connection from server and print messages
func runConnectBack(serverIP net.IP) {
	// Find local IP (which interface would be used to reach server)
	myIP := localIPv4For(serverIP)
	if myIP == nil {
		log.Println("Connect-back: could not determine local IP")
		return
	}

	ln, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: myIP, Port: 0})
	if err != nil {
		log.Println("Connect-back: ListenTCP failed:", err)
		return
	}
	fmt.Println("Connect-back listening on", ln.Addr().String())

	// Accept in its own goroutine
	go func() {
		defer ln.Close()

		_ = ln.SetDeadline(time.Now().Add(10 * time.Second)) // just to avoid waiting forever if something is wrong
		backConn, err := ln.AcceptTCP()
		if err != nil {
			log.Println("Connect-back: AcceptTCP error:", err)
			return
		}
		defer backConn.Close()

		r := bufio.NewReader(backConn)
		for {
			msg, err := r.ReadBytes(0)
			if err != nil {
				log.Println("Connect-back: read error:", err)
				return
			}
			fmt.Println("Connect-back recv:", string(bytes.TrimRight(msg, "\x00")))
		}
	}()

	// Send the connect-back command to server over the \0 TCP port
	myPort := ln.Addr().(*net.TCPAddr).Port
	cmdConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: serverIP, Port: tcpZeroPort})
	if err != nil {
		log.Println("Connect-back: DialTCP to command port failed:", err)
		return
	}
	defer cmdConn.Close()

	cmd := fmt.Sprintf("Connect to: %s:%d\x00", myIP.String(), myPort)
	_, err = cmdConn.Write([]byte(cmd))
	if err != nil {
		log.Println("Connect-back: sending command failed:", err)
		return
	}
	fmt.Println("Connect-back command sent:", strings.TrimRight(cmd, "\x00"))
}

// Determine local IPv4 address used to reach serverIP (no hardcoding).
func localIPv4For(serverIP net.IP) net.IP {
	c, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: serverIP, Port: udpDiscoveryPort})
	if err != nil {
		return nil
	}
	defer c.Close()

	ua, ok := c.LocalAddr().(*net.UDPAddr)
	if !ok || ua.IP == nil {
		return nil
	}
	return ua.IP
}

func main() {
	serverIP := discoverServerIP()
	fmt.Println("Server IP:", serverIP)

	// Start both TCP variants (comment out if you only want one)
	go runTCPZero(serverIP)
	go runTCPFixed(serverIP)

	// Connect-back (bonus). Uncomment when tcp0 works.
	go runConnectBack(serverIP)

	select {} // keep main alive
}
