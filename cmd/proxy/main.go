package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"time"

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
	go func() {
		return
		l, err := net.FileListener(f)
		fmt.Println("opened file listener", l)
		if err != nil {
			fmt.Println("failed to retrieve listener", err)
			return
		}

		for {
			fmt.Println("waiting to accept")
			// conn, err := l.Accept()
			fmt.Println(l.Addr())
			fmt.Printf("addr % #v\n", l.Addr())
			conn, err := net.ListenUnixgram("unixgram", l.Addr().(*net.UnixAddr))
			fmt.Println("accepted conn")
			if err != nil {
				fmt.Println("err accepting conn", err)
				return
			}
			bs := make([]byte, 512)
			fmt.Println("reading from conn")
			n, err := conn.Read(bs)
			if err != nil {
				fmt.Println("err reading conn", err)
				continue
			}
			fmt.Println("read", n, "bytes")
			fmt.Println(string(bs[:n]))
		}
	}()
	// time.Sleep(1 * time.Minute)
	// return
	upstream, err := net.FileConn(f)
	if err != nil {
		// log.Fatal(err)
	}
	// uf, _ := upstream.(*net.UnixConn).File()
	// fmt.Printf("% #v\n", pretty.Formatter(uf))

	if _, err = os.Stat(socks.PATH); err == nil {
		os.Remove(socks.PATH)
	}

	defer func() {
		os.Remove(socks.PATH)
		fmt.Println("removed", socks.PATH)
	}()

	addr, err := net.ResolveUnixAddr("unixgram", socks.PATH)
	conn, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {

			b, obb := make([]byte, 512), make([]byte, 512)
			n, oobn, flags, addr, err := conn.ReadMsgUnix(b, obb)
			fmt.Println(n, oobn, flags, addr, err)

		}
	}()
	ff, err := conn.File()
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			_ = sig
			fmt.Println("shutdown requested")
			log.Fatal("done")
		}
	}()

	fmt.Println("listening on", socks.PATH)
	if true {
		go func() {
			s := "/usr/local/bin/ocproxy"
			cmd := exec.Command(s, "-D 11080", "-v")
			cmd.Env = []string{
				fmt.Sprintf("VPNFD=%d", ff.Fd()),
				"INTERNAL_IP4_DNS=144.20.190.70",
				"INTERNAL_IP4_ADDRESS=10.154.167.196",
				"INTERNAL_IP4_MTU=1300",
			}
			cmd.ExtraFiles = []*os.File{ff}
			fmt.Println("staring with", ff.Fd())
			bs, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(err)
			}
			fmt.Println("cmd output", string(bs))
		}()
	}
	time.Sleep(1 * time.Minute)
	return
	l, err := net.Listen("unix", socks.PATH)
	if err != nil {
		log.Fatal(err)
	}
	ul := l.(*net.UnixListener)

	// Start socks proxy server. Accepting incoming requests
	// to forward to the vpnfd sock
	serve(upstream, ul)

}

func listenAndWrite(upstream net.Conn, in *net.UnixConn) {
	bs := make([]byte, 512)
	fmt.Println("waiting to read")
	var wrotesomething bool
	_ = wrotesomething
	go func() {
		fmt.Println("read routine")
		var ups []byte
		for {
			fmt.Println("reading from upstream")
			n, err := upstream.Read(ups)
			if err != nil {
				fmt.Println("error reading from upstream", err)
				break
			}
			if n == 0 {
				fmt.Println("reading 0 bytes")
				break
			}
			fmt.Printf("read %d bytes: err\n", n, err)
		}
	}()

	for {
		n, err := in.Read(bs)
		if err != nil {
			fmt.Println("in read err, breaking", err)
			break
		}
		if n == 0 {
			fmt.Println("read nothing breaking")
			break
		}
		if n > 0 {
			fmt.Printf("found %d bytes\n", n)
			fmt.Println("writing to upstream")
			wrotesomething = true
			nn, err := upstream.Write(bs[:n])
			if err != nil {
				fmt.Println("error writing to upstream",
					err)
			}
			fmt.Println("wrote", nn, "bytes")
		}
	}
}

func serve(upstream net.Conn, l *net.UnixListener) {
	for {
		fmt.Println("waiting to accept")
		conn, err := l.AcceptUnix()
		if err != nil {
			fmt.Println("error accepting conn", err)
			break
		}
		fmt.Println("accepted conn", conn)
		// these leak
		go listenAndWrite(upstream, conn)
	}
}
