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
	"sync"

	"golang.org/x/net/proxy" // "github.com/golang/net/proxy"

	"github.com/armon/go-socks5"
	"github.com/hashicorp/hcl"
)

func (o *res) dump() {
	for k, v := range o.names {
		fmt.Println("k~>", k, "v~>", v)
	}
}

func (o *res) dialer(n, a string) (net.Conn, error) {
	// port is bogus from this, lookup the port from the resolver
	h, notthisport, err := net.SplitHostPort(a)
	_ = notthisport
	fmt.Println("dialer", n, a)
	if err != nil {
		fmt.Println("failed to split hostport", a)
	}
	// Does a lookup locally for the fqdn
	// names, err := net.LookupAddr(h)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	name := o.Lookup(h)
	fmt.Println("reverse", h, "name", name)
	if o.checkName(name) {
		fmt.Println("dialing", n, name+":"+notthisport)
		return forward.Dial(n, name+":"+notthisport)
	} else {
		fmt.Println("direct", name)
	}

	// Use standard resolver
	return net.Dial(n, a)
}

var forward proxy.Dialer

// FileConfig defines the allowed parameters read from a file.
type FileConfig struct {
	Forward string   // addr of proxy aware socks5 server
	Listen  string   // addr to listen on
	Allow   []string // Slice of patterns to forward
	// TODO: Disallow []string
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	bs, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, "proxy.cfg"))
	if err != nil {
		log.Fatal(err)
	}

	cfg := FileConfig{}
	err = hcl.Decode(&cfg, string(bs))
	if err != nil {
		log.Fatal(err)
	}

	fURL, err := url.Parse("socks5://" + cfg.Forward)
	if err != nil {
		log.Fatal(err)
	}
	forward, err = proxy.FromURL(fURL, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}
	r := &res{
		rules: cfg.Allow,
	}
	r.init()

	// Create a SOCKS5 server
	conf := &socks5.Config{}
	conf.Dial = r.dialer
	conf.Resolver = r
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
	cache map[string]struct{}
}

func (r *res) init() {
	r.names = make(map[string]net.Addr)
	r.cache = make(map[string]struct{})
}

func (r *res) Lookup(host string) string {
	host = host + ":0"
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

var empty = net.IP{}

var cMu sync.Mutex
var counter = net.IP{0, 0, 0, 0}

func getCounter() net.IP {
	cMu.Lock()
	inc()
	n := net.IP{0, 0, 0, 0}
	copy(n, counter)
	cMu.Unlock()
	return n
}

func inc() {
	carry := true
	fmt.Printf("% #v\n", counter)
	for i := 3; carry; i-- {
		if counter[i] < 254 {
			carry = false
		}
		counter[i]++
	}
}

func (r *res) Resolve(name string) (net.IP, error) {
	fmt.Println("resolving", name)
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
		// Proxy is required to resolve this IP, pass a code
		// so the dialer knows this requires resolution
		ip = getCounter()
	} else {
		ip, err = r.def.Resolve(name)
		if err != nil {
			log.Fatal("failed to resolve addr", err)
		}
	}
	addr = &net.TCPAddr{IP: ip}
	r.mu.Lock()
	fmt.Println("storing", name, addr)
	r.names[name] = addr
	r.mu.Unlock()
	if err != nil {
		log.Fatal("error in resolve", err)
	}
	return ip, err
}
