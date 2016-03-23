package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/armon/go-socks5"
	"github.com/drewwells/proxy"
	"github.com/hashicorp/hcl"
	"golang.org/x/net/proxy"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	bs, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, "proxy.cfg"))
	if err != nil {
		log.Fatal(err)
	}

	cfg := socks.FileConfig{}
	err = hcl.Decode(&cfg, string(bs))
	if err != nil {
		log.Fatal(err)
	}

	r := &socks.Res{}
	env := socks.GatherEnv()
	// env.VPNFD = 9
	var conn net.Conn
	if env.VPNFD > 0 {
		fmt.Println(os.Environ())
		fmt.Printf("% #v\n", env)
		f := os.NewFile(env.VPNFD, "mysocket")
		conn, err = net.FileConn(f)
		fmt.Printf("conn % #v\n", conn)
		if err != nil {
			log.Fatal(err)
		}
	}

	if conn != nil {
		fmt.Println("local", conn.LocalAddr())
		fmt.Println("remote", conn.RemoteAddr())
		r.SetConn(conn)
	} else if strings.HasPrefix(cfg.Forward, "/") {
		addr := &net.UnixAddr{
			Net:  "unix",
			Name: cfg.Forward,
		}
		fmt.Println("dialunix")
		c, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			panic(err)
		}
		fmt.Println("setconn")
		r.SetConn(c)
	} else {
		fURL, err := url.Parse("socks5://" + cfg.Forward)
		if err != nil {
			log.Fatal(err)
		}
		forward, err := proxy.FromURL(fURL, proxy.Direct)
		if err != nil {
			log.Fatal(err)
		}
		r.SetForward(forward)
	}

	// r.SetConn(conn)
	r.SetWhitelist(cfg.Allow)
	r.SetBlacklist(cfg.Block)
	r.Init()

	// Create a SOCKS5 server
	conf := &socks5.Config{
		Dial:     r.Dialer,
		Resolver: r,
		Logger:   log.New(os.Stderr, "", 0),
	}

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening  on", cfg.Listen)
	fmt.Println("Forwarding to", cfg.Forward)
	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", cfg.Listen); err != nil {
		panic(err)
	}

}
