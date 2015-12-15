package dockerclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	default:
		httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(proto, addr, timeout)
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				// Set TCP user timeout. Sender breaks TCP connection
				// if packets are not acknowledged after 20 seconds. This is a
				// relatively new TCP option to improve dead peer detection.
				// Do not fail newHTTPClient if OS doesn's support it.
				SetTCPUserTimeout(tcpConn, 20*time.Second)
			}
			return conn, err
		}
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}
