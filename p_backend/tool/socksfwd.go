// tool/socksfwd.go - TCP port forwarder through SOCKS5 proxy.
// Usage: go run tool/socksfwd.go [proxy] [remote] [local_port]
// Example: go run tool/socksfwd.go 127.0.0.1:7890 db.example.com:5432 15432
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func dialSocks5(proxyAddr, target string) (net.Conn, error) {
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte{0x05, 0x01, 0x00})
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		conn.Close()
		return nil, err
	}
	host, portStr, _ := net.SplitHostPort(target)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(host))}
	req = append(req, []byte(host)...)
	pb := make([]byte, 2)
	binary.BigEndian.PutUint16(pb, uint16(port))
	req = append(req, pb...)
	conn.Write(req)
	resp := make([]byte, 10)
	if _, err := io.ReadFull(conn, resp); err != nil {
		conn.Close()
		return nil, err
	}
	if resp[1] != 0x00 {
		conn.Close()
		return nil, fmt.Errorf("socks5 connect failed: code %d", resp[1])
	}
	return conn, nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run tool/socksfwd.go <proxy> <remote> <local_port>")
		fmt.Println("Example: go run tool/socksfwd.go 127.0.0.1:7890 db.example.com:5432 15432")
		os.Exit(1)
	}
	proxyAddr := os.Args[1]
	remoteAddr := os.Args[2]
	localPort := os.Args[3]

	ln, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("forwarding 127.0.0.1:%s -> %s via %s", localPort, remoteAddr, proxyAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			remote, err := dialSocks5(proxyAddr, remoteAddr)
			if err != nil {
				log.Println("dial:", err)
				c.Close()
				return
			}
			go io.Copy(remote, c)
			io.Copy(c, remote)
			c.Close()
			remote.Close()
		}(conn)
	}
}
