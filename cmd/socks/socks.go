package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/armon/go-socks5"
	"github.com/drewwells/socks"
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
	// env := gatherenv()
	// f := os.NewFile(env.VPNFD, "mysocket")
	// conn, err := net.FileConn(f)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fURL, err := url.Parse("socks5://" + cfg.Forward)
	if err != nil {
		log.Fatal(err)
	}
	forward, err := proxy.FromURL(fURL, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}
	// r.SetConn(conn)
	r.SetRules(cfg.Allow)
	r.SetForward(forward)
	r.Init()

	// Create a SOCKS5 server
	conf := &socks5.Config{}
	conf.Dial = r.Dialer
	conf.Resolver = r
	conf.Logger = log.New(os.Stderr, "", 0)

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", cfg.Listen); err != nil {
		panic(err)
	}

}
