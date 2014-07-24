package dockerclient

type ContainerConfig struct {
	Hostname        string
	Domainname      string
	User            string
	Memory          int
	MemorySwap      int
	CpuShares       int
	Cpuset          string
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	PortSpecs       []string
	ExposedPorts    map[string]struct{}
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             []string
	Image           string
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	OnBuild         []string
}

type HostConfig struct {
	Binds           []string
	ContainerIDFile string
	LxcConf         []map[string]string
	Privileged      bool
	PortBindings    map[string][]PortBinding
	Links           []string
	PublishAllPorts bool
	Dns             []string
	DnsSearch       []string
	VolumesFrom     []string
	NetworkMode     string
}

type PortBinding struct {
	HostIp   string
	HostPort string
}

type ContainerInfo struct {
	Id      string
	Created string
	Path    string
	Name    string
	Args    []string
	Config  *ContainerConfig
	State   struct {
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

type Port struct {
	PrivatePort int
	PublicPort  int
	Type        string
}

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

type Event struct {
	Id     string
	Status string
	From   string
	Time   int
}

type Version struct {
	Version   string
	GitCommit string
	GoVersion string
}

type RespContainersCreate struct {
	Id       string
	Warnings []string
}

type Image struct {
	Created     int64
	Id          string
	ParentId    string
	RepoTags    []string
	Size        int64
	VirtualSize int64
}
