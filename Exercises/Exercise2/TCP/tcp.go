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

var (
	fixedAddr = &net.TCPAddr{IP: net.ParseIP("10.100.23.23"), Port: 34933}
	zeroAddr = &net.TCPAddr{IP: net.ParseIP("10.100.23.23"), Port: 33546}
)

func pad1024(msg string) []byte {
	b := make([]byte, 1024)
	copy(b, []byte(msg+"\000"))
	return b
}

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

func tcpFixed1024(serverIP net.IP) {
	conn, err := net.Dial("tcp4", net.JoinHostPort(serverIP.String(), "34933"))
	if err != nil { log.Fatal(err) }
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err = io.ReadFull(conn, buf)
	if err != nil { log.Fatal(err) }
	fmt.Println("TCP fixed welcome:", string(bytes.TrimRight(buf, "\x00")))

	_, err = conn.Write(pad1024("hello fixed"))
	if err != nil { log.Fatal(err) }

	_, err = io.ReadFull(conn, buf)
	if err != nil { log.Fatal(err) }
	fmt.Println("TCP fixed echo:", string(bytes.TrimRight(buf, "\x00")))
}

func tcpConnectBack(serverIP net.IP, myIP net.IP) {
	listener, err := net.ListenTCP("tcp4", fixedAddr)
	if err != nil { log.Fatal(err) }
	defer listener.Close()

	myPort := fixedAddr.Port

	fmt.Println("Listening to fixedAddr")

	conn, err := net.Dial("tcp4", net.JoinHostPort(serverIP.String(), string(myPort)))
	if err != nil{log.Fatal(err)}
	defer conn.Close()

	cmd := fmt.Sprintf("Connect to: %s:%d", myIP.String(), myPort)

	_,err = conn.Write(([]byte(cmd)))
	if err != nil{log.Fatal(err)}

	backConn, err := listener.AcceptTCP()
	if err != nil {log.Fatal(err)}
	defer backConn.Close()

	r := bufio.NewReader(backConn)
	msg, err := r.ReadBytes(0)
	if err != nil{log.Fatal(err)}
	fmt.Println("Connectback recv", string(bytes.TrimRight(msg,"\x00")))
}

func main() {
	serverIP := getServerIP()
	fmt.Println("Server IP:", serverIP)

	go tcpFixed1024(serverIP)

	//go tcpConnectBack(serverIP, fixedAddr.IP)

	select {}
}