// dockerclient
// For the full copyright and license information, please view the LICENSE file.

// This file contains dockerclient types.

package dockerclient

import (
	"io"
)

// ContainerConfig implements a container configuration.
type ContainerConfig struct {
	Hostname        string
	Domainname      string
	User            string
	Memory          int
	MemorySwap      int
	CpuShares       int
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             []string
	Dns             []string
	Image           string
	VolumesFrom     string
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
}

// HostConfig implements a host configuration.
type HostConfig struct {
	Binds           []string
	ContainerIDFile string
	LxcConf         []map[string]string
	Privileged      bool
	PortBindings    map[string][]PortBinding
	Links           []string
	PublishAllPorts bool
}

// PortBinding implements a port binding.
type PortBinding struct {
	HostIp   string
	HostPort string
}

// ContainerInfo implements a container information.
type ContainerInfo struct {
	Id     string
	Create string
	Path   string
	Args   []string
	Config *ContainerConfig
	State  struct {
		Running   bool
		Pid       int
		ExitCode  int
		StartedAt string
		Ghost     bool
	}
	Image           string
	NetworkSettings struct {
		IpAddress   string
		IpPrefixLen int
		Gateway     string
		Bridge      string
		Ports       map[string][]PortBinding
	}
	SysInitPath    string
	ResolvConfPath string
	Volumes        map[string]string
	HostConfig     *HostConfig
}

// Port implements a port.
type Port struct {
	PrivatePort int
	PublicPort  int
	Type        string
}

// Container implements a container.
type Container struct {
	Id         string
	Names      []string
	Image      string
	Command    string
	Created    int
	Status     string
	Ports      []Port
	SizeRw     int
	SizeRootFs int
}

// Event implements a event.
type Event struct {
	Id     string
	Status string
	From   string
	Time   int
}

// Version implements a version information.
type Version struct {
	Version   string
	GitCommit string
	GoVersion string
}

// Attach implements attach protocol.
type Attach struct {
	Tty       bool
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	StdinPipe io.Reader
}

// RespContainersCreate implements the response of a create container request.
type RespContainersCreate struct {
	Id       string
	Warnings []string
}
