package socks

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy" // "github.com/golang/net/proxy"

	"github.com/armon/go-socks5"
)

const PATH = "/tmp/proxy.sock"

func (o *Res) dump() {
	for k, v := range o.names {
		fmt.Println("k~>", k, "v~>", v)
	}
}

func (o *Res) Dialer(n, a string) (net.Conn, error) {
	// port is bogus from this, lookup the port from the resolver
	h, notthisport, err := net.SplitHostPort(a)
	_ = notthisport
	if err != nil {
		fmt.Println("failed to split hostport", a)
	}
	// Does a lookup locally for the fqdn
	// names, err := net.LookupAddr(h)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	name := o.Lookup(h)
	if o.checkName(name) {
		// return o.conn, nil
		// fmt.Println("proxy", name+":"+notthisport)
		return o.forward.Dial(n, name+":"+notthisport)
	} else {
		// fmt.Println("direct", name)
	}

	// Use standard resolver
	return net.Dial(n, a)
}

// FileConfig defines the allowed parameters read from a file.
type FileConfig struct {
	Forward string   // addr of proxy aware socks5 server
	Listen  string   // addr to listen on
	Allow   []string // Slice of patterns to forward
	// TODO: Disallow []string
}

var _ socks5.NameResolver = &Res{}

type Res struct {
	conn    net.Conn
	forward proxy.Dialer

	def   socks5.DNSResolver
	rules []string
	mu    sync.RWMutex
	names map[string]net.Addr // host(ip:port) -> fqdn
	cache map[string]struct{}
}

func (r *Res) SetRules(rules []string) {
	r.rules = rules
}

func (r *Res) SetConn(c net.Conn) {
	r.conn = c
}

func (r *Res) SetForward(d proxy.Dialer) {
	r.forward = d
}

func (r *Res) Init() {
	r.names = make(map[string]net.Addr)
	r.cache = make(map[string]struct{})
	go func() {
		for {
			time.Sleep(time.Second * 30)
			fmt.Printf("writes: %d\nreads:  %d\n", stored, lookup)
		}
	}()
}

var lookup int

func (r *Res) Lookup(host string) string {
	host = host + ":0"
	lookup++
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
func (r *Res) checkName(name string) bool {
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
	for i := 3; carry; i-- {
		if counter[i] < 254 {
			carry = false
		}
		counter[i]++
	}
}

func (r *Res) Resolve(name string) (net.IP, error) {
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
			log.Println("failed to resolve addr", err)
		}
	}
	addr = &net.TCPAddr{IP: ip}
	r.mu.Lock()
	stored++
	// fmt.Println("storing", name, addr)
	r.names[name] = addr
	r.mu.Unlock()
	if err != nil {
		log.Fatal("error in resolve", err)
	}
	return ip, err
}

var stored int
