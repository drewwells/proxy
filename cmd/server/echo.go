package main

import (
	"fmt"
	"log"
	"net"

	"github.com/drewwells/socks"
)

func echoServer(c net.Conn) {
	for {
		buf := make([]byte, 512)
		fmt.Println("waiting to read")
		nr, err := c.Read(buf)
		fmt.Println("read", nr, err)
		if err != nil {
			fmt.Println("exit")
			return
		}

		data := buf[0:nr]
		println("Server got:", string(data))
		_, err = c.Write(data)
		if err != nil {
			log.Fatal("Write: ", err)
		}
	}
}

func main() {
	l, err := net.Listen("unix", socks.PATH)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		fmt.Println("conn accepted")
		go echoServer(conn)
	}
}
