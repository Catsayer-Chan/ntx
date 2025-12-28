package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	target := "8.8.8.8"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	fmt.Printf("Testing ICMP ping to %s\n", target)
	fmt.Printf("OS: %s\n", runtime.GOOS)

	// 解析目标地址
	dst, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		log.Fatalf("ResolveIPAddr failed: %v", err)
	}
	fmt.Printf("Resolved to: %s\n", dst.String())

	// 尝试 udp4 with 0.0.0.0
	fmt.Println("\n=== Trying udp4 with 0.0.0.0 ===")
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Printf("udp4 (0.0.0.0) failed: %v", err)
	} else {
		fmt.Println("udp4 (0.0.0.0) connection created successfully")
		testPing(conn, dst)
		conn.Close()
	}

	// 尝试 udp4 without address
	fmt.Println("\n=== Trying udp4 with empty address ===")
	conn, err = icmp.ListenPacket("udp4", "")
	if err != nil {
		log.Printf("udp4 (empty) failed: %v", err)
	} else {
		fmt.Println("udp4 (empty) connection created successfully")
		testPing(conn, dst)
		conn.Close()
	}

	// 尝试 ip4:icmp
	fmt.Println("\n=== Trying ip4:icmp ===")
	conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Printf("ip4:icmp failed: %v", err)
	} else {
		fmt.Println("ip4:icmp connection created successfully")
		testPing(conn, dst)
		conn.Close()
	}
}

func testPing(conn *icmp.PacketConn, dst *net.IPAddr) {
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("HELLO"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		log.Printf("Marshal failed: %v", err)
		return
	}

	// 设置超时
	deadline := time.Now().Add(5 * time.Second)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	// 发送
	start := time.Now()
	n, err := conn.WriteTo(msgBytes, dst)
	if err != nil {
		log.Printf("WriteTo failed: %v", err)
		return
	}
	fmt.Printf("Sent %d bytes\n", n)

	// 接收
	recvBuf := make([]byte, 1500)
	n, peer, err := conn.ReadFrom(recvBuf)
	if err != nil {
		log.Printf("ReadFrom failed: %v", err)
		return
	}

	rtt := time.Since(start)
	fmt.Printf("Received %d bytes from %s in %v\n", n, peer.String(), rtt)

	// 解析响应
	rm, err := icmp.ParseMessage(1, recvBuf[:n])
	if err != nil {
		log.Printf("ParseMessage failed: %v", err)
		return
	}

	fmt.Printf("Response type: %v, Code: %d\n", rm.Type, rm.Code)
}
