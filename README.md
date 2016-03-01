Proxy starts a socks5 proxy that listens for traffic. URLs passed to it are compared against the Allow list. If a match is found, that traffic is directed to the Forward socks5 proxy for passing to a VPN.

Proxy is useful when combined with ocproxy and openclient. See gist here for more info: https://gist.github.com/drewwells/a254588e52766104ec9c

Place your config file at ~/proxy.cfg (not configurable).

Future:
direct support for openclient, replacing the need for ocproxy
