// +build windows

package dockerclient

// #include <Ws2tcpip.h>
import "C"

import (
	"net"
	"os"
	"syscall"
	"time"
)

// SetTCPUserTimeout sets TCP_MAXRT in Windows
func SetTCPUserTimeout(conn *net.TCPConn, uto time.Duration) error {
	f, err := conn.File()
	if err != nil {
		return err
	}
	defer f.Close()

	// TCP_MAXRT in Windows is set as seconds
	secs := int(uto.Nanoseconds() / 1e9)

	// from MSDN, TCP_MAXRT is supported since Windows Vista
	return os.NewSyscallError("setsockopt", syscall.SetsockoptInt(int(f.Fd()), syscall.SOL_TCP, C.TCP_MAXRT, secs))
}
