package dockerclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
)

func newHTTPClient(u *url.URL, tlsConfig *tls.Config) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	if u.Scheme == "unix" {
		socketPath := u.Path
		unixDial := func(proto string, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
	}
	u.Path = ""
	return &http.Client{Transport: httpTransport}
}
