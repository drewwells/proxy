Proxy starts a socks5 proxy that listens for traffic. URLs passed to it are compared against the Allow list. If a match is found, that traffic is directed to the Forward socks5 proxy for passing to a VPN. This is to replace tools like foxyproxy that work only in the browser.

#### How to use
Setup a socks proxy to your VPN. Gist provided to do this with openclient and ocproxy: https://gist.github.com/drewwells/a254588e52766104ec9c

By default, ocproxy runs on port 11080 on localhost. So we will configure proxy to forward requests to that. Add a file at `~/proxy.cfg`.

    ~/proxy.cfg

``` hcl
Forward = "127.0.0.1:11080"
Listen = "127.0.0.1:7999"
Allow = ["mycompany.org"]
```

Now configure your system to use a SOCKS5 proxy at `127.0.0.1:7999`. Proxy will inspect the traffic and determine if the request belongs on the VPN or not. For this configuration, any requests matching `mycompany.org`

Place your config file at ~/proxy.cfg example at example.proxy.cfg).

#### To install

    go install github.com/drewwells/proxy/cmd/socks

Future:
Communicate directly with openclient, eliminating need for ocproxy

Stuck on this: https://groups.google.com/forum/#!topic/golang-nuts/ombu872TNFY
