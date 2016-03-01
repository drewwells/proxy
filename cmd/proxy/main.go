package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const path = "/tmp/proxy.sock"

type Config struct {
	// Environment variables provided by openconnect
	IP4InternalAddress string
	IP4InternalMTU     string
	VPNFD              uintptr
	IP4InternalDNS     []string
	CiscoDefDomain     []string
}

func gatherenv() *Config {
	cfg := &Config{}
	cfg.IP4InternalAddress = os.Getenv("INTERNAL_IP4_ADDRESS")
	cfg.IP4InternalMTU = os.Getenv("INTERNAL_IP4_MTU")
	domains := os.Getenv("CISCO_DEF_DOMAIN")
	cfg.CiscoDefDomain = strings.Split(domains, ", ")
	dnses := os.Getenv("INTERNAL_IP4_DNS") // space sparated values
	cfg.IP4InternalDNS = strings.Split(dnses, " ")

	svpnfd := os.Getenv("VPNFD")
	ivpnfd, err := strconv.Atoi(svpnfd)
	if err != nil {
		log.Fatal("invalid fd", err)
	}
	cfg.VPNFD = uintptr(ivpnfd)
	return cfg
}

func main() {
	log.SetFlags(log.Llongfile)

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
	cfg := gatherenv()
	fmt.Printf("% #v\n", cfg)
	f := os.NewFile(cfg.VPNFD, "mysocket")
	conn, err := net.FileConn(f)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("conn % #v\n", conn)
	n, err := conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	log.Println("write", n, err)

	go func() {
		for {
			result, err := ioutil.ReadAll(conn)
			log.Println("readall", err, result)
		}
	}()

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
