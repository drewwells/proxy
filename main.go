package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/proxy" // "github.com/golang/net/proxy"

	"github.com/armon/go-socks5"
)

type opts struct {
	// these are contains filters ie. goo matches google.com or good.com
	resolver *res
}

func (o *opts) dialer(n, a string) (net.Conn, error) {
	// port is bogus from this, lookup the port from the resolver
	h, p, err := net.SplitHostPort(a)
	fmt.Println("split", h, p)
	if err != nil {
		fmt.Println("failed to split hostport", a)
	}
	// Does a lookup locally for the fqdn
	// names, err := net.LookupAddr(h)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	fmt.Println("dialer lookup")
	name := o.resolver.Lookup(h)
	fmt.Println("reverse", h, name)
	if o.resolver.checkName(name) {
		fmt.Println("proxying", n, a)
		dialer, err := proxy.SOCKS5(n, a, nil, forward)
		if err != nil {
			fmt.Println("dialer error")
			return nil, err
		}

		return dialer.Dial(n, a)
	} else {
		fmt.Println("direct", name)
	}

	// Use standard resolver
	return net.Dial(n, a)
}

var forward proxy.Dialer

func main() {
	fURL, err := url.Parse("socks5://localhost:11080")
	if err != nil {
		log.Fatal(err)
	}
	forward, err = proxy.FromURL(fURL, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}
	r := &res{
		rules: []string{"oracle"},
	}
	r.init()
	o := &opts{
		resolver: r,
	}

	// Create a SOCKS5 server
	conf := &socks5.Config{}
	conf.Dial = o.dialer
	conf.Resolver = o.resolver
	conf.Logger = log.New(os.Stderr, "", 0)

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", "127.0.0.1:8000"); err != nil {
		panic(err)
	}
}

var _ socks5.NameResolver = &res{}

type res struct {
	def   socks5.DNSResolver
	rules []string
	mu    sync.RWMutex
	names map[string]net.Addr // host(ip:port) -> fqdn
}

func (r *res) Lookup(host string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for k, v := range r.names {
		if v.String() == host {
			return k
		}
	}
	return ""
}

// checknames compares the resolved addresses against the
// the whitelist of URLs
func (r *res) checkName(name string) bool {
	for _, rule := range r.rules {
		if strings.Contains(name, rule) {
			return true
		}
	}
	return false
}

func (r *res) init() {
	r.names = make(map[string]net.Addr)
}

func ipFromAddr(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPAddr:
		return v.IP
	case *net.TCPAddr:
		return v.IP
	default:
		panic(fmt.Errorf("unsupported type %T", v))
	}
}

func (r *res) Resolve(name string) (net.IP, error) {
	r.mu.RLock()
	addr, ok := r.names[name]
	r.mu.RUnlock()
	if ok {
		return ipFromAddr(addr), nil
	}

	var (
		err error
		ip  net.IP
	)
	// Resolve this name with proxy
	if r.checkName(name) {
		h, p, err := net.SplitHostPort(name)
		fmt.Println(h, p)
		c, err := forward.Dial("tcp", name+":80")
		if err != nil {
			log.Fatal(err)
		}
		addr = c.LocalAddr()
		ip = ipFromAddr(addr)
		fmt.Printf("proxy found %s: % #v\n", ip, addr)
	} else {
		ip, err = r.def.Resolve(name)
		if err != nil {
			log.Fatal("failed to resolve addr", err)
		}
	}
	r.mu.Lock()
	fmt.Println("storing", name, addr)
	r.names[name] = addr
	r.mu.Unlock()
	// Builds a local cache of IP to address for use by the dialer
	return ip, err
}
