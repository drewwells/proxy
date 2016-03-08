package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/drewwells/proxy"
)

func main() {
	log.SetFlags(log.Llongfile)
	cfg := socks.GatherEnv()
	fmt.Printf("% #v\n", cfg)
	cfg.VPNFD = 9
	if cfg.VPNFD == 0 {
		log.Println("vpnfd missing")
		for _, env := range os.Environ() {
			fmt.Println(env)
		}
		log.Fatal("")
	}

	f := os.NewFile(cfg.VPNFD, "mysocket")
	upstream, err := net.FileConn(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("local % #v\n", upstream.LocalAddr())
	fmt.Printf("remot % #v\n", upstream.RemoteAddr())

	if _, err = os.Stat(socks.PATH); err == nil {
		os.Remove(socks.PATH)
	}

	defer func() {
		os.Remove(socks.PATH)
		fmt.Println("removed", socks.PATH)
	}()

	l, err := net.Listen("unix", socks.PATH)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("listening on", socks.PATH)
	// Start socks proxy server. Accepting incoming requests
	// to forward to the vpnfd sock
	serve(upstream, l)

}

func listenAndWrite(upstream, in net.Conn) {
	bs := make([]byte, 512)
	fmt.Println("waiting to read")
	var readsomething bool
	for {
		n, err := in.Read(bs)
		if err != nil {
			break
		}
		if n > 0 {
			fmt.Printf("found %d bytes\n", n)
			fmt.Println(string(bs))
			readsomething = true
			_, err := upstream.Write(bs[:n])
			if err != nil {
				fmt.Println("error writing to upstream",
					err)
			}
		}
	}
	if readsomething {
		var ups []byte
		for {
			n, err := upstream.Read(ups)
			if err != nil {
				fmt.Println("error reading from upstream", err)
				break
			}
			if n == 0 {
				fmt.Println("reading 0 bytes")
				break
			}
			fmt.Printf("reading %d bytes: err\n", n, err)
		}
	}
}

func serve(upstream net.Conn, l net.Listener) {
	for {
		fmt.Println("waiting to accept")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error accepting conn", err)
			break
		}
		// these leak
		go listenAndWrite(upstream, conn)
	}
}
