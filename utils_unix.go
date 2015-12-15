// +build !windows

package dockerclient

// #include <netinet/tcp.h>
import "C"

import (
	"net"
	"os"
	"syscall"
	"time"
)

// SetTCPUserTimeout sets TCP_USER_TIMEOUT according to RFC5842
func SetTCPUserTimeout(conn *net.TCPConn, uto time.Duration) error {
	f, err := conn.File()
	if err != nil {
		return err
	}
	defer f.Close()

	msecs := int(uto.Nanoseconds() / 1e6)
	return os.NewSyscallError("setsockopt", syscall.SetsockoptInt(int(f.Fd()), syscall.SOL_TCP, C.TCP_USER_TIMEOUT, msecs))
}
