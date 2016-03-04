package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

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
		// log.Fatal(err)
	}

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			_ = sig
			fmt.Println("shutdown requested")
			l.Close()
		}
	}()

	fmt.Println("listening on", socks.PATH)
	// Start socks proxy server. Accepting incoming requests
	// to forward to the vpnfd sock
	serve(upstream, l)

	// in, err := net.Listen("unix", path)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer fmt.Println("closing socket")
	// defer in.Close()
	// fmt.Printf("% #v\n", in)

	// ss := os.Environ()
	// for _, s := range ss {
	// 	fmt.Println(s)
	// }

	// log.Printf("conn % #v\n", conn)
	// n, err := conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	// log.Println("write", n, err)

	// go func() {
	// 	for {
	// 		result, err := ioutil.ReadAll(conn)
	// 		log.Println("readall", err, result)
	// 	}
	// }()

	// laddr, err := net.ResolveUnixAddr("unix", "")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// raddr, err := net.ResolveUnixAddr("unix", path)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// conn, err := net.DialUnix("unix", laddr, raddr)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("getting fd", vpnfd)
	// f, err := fd.Get(conn, vpnfd, []string{"duh"})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("f[0] % #v\n", f[0])
	// names, err := f[0].Readdirnames(9)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("names", names)
	// _, err = f[0].WriteString("Stuff happening now")
	// fmt.Println(err)
	// // f := os.NewFile(vpnfd, "tunfile")
	// fmt.Printf("% #v\n", f)
	// log.Println("VPNFD:", vpnfd)
	// log.Fatal(os.Args)
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
