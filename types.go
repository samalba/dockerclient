package dockerclient

import (
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/pkg/units"
)

type ContainerConfig struct {
	Hostname        string                    `json:"Hostname,omitempty" yaml:"Hostname,omitempty"`
	Domainname      string                    `json:"Domainname,omitempty" yaml:"Domainname,omitempty"`
	User            string                    `json:"User,omitempty" yaml:"User,omitempty"`
	AttachStdin     bool                      `json:"AttachStdin,omitempty" yaml:"AttachStdin,omitempty"`
	AttachStdout    bool                      `json:"AttachStdout,omitempty" yaml:"AttachStdout,omitempty"`
	AttachStderr    bool                      `json:"AttachStderr,omitempty" yaml:"AttachStderr,omitempty"`
	ExposedPorts    map[string]struct{}       `json:"ExposedPorts,omitempty" yaml:"ExposedPorts,omitempty"`
	Tty             bool                      `json:"Tty,omitempty" yaml:"Tty,omitempty"`
	OpenStdin       bool                      `json:"OpenStdin,omitempty" yaml:"OpenStdin,omitempty"`
	StdinOnce       bool                      `json:"StdinOnce,omitempty" yaml:"StdinOnce,omitempty"`
	Env             []string                  `json:"Env,omitempty" yaml:"Env,omitempty"`
	Cmd             []string                  `json:"Cmd" yaml:"Cmd"`
	Image           string                    `json:"Image,omitempty" yaml:"Image,omitempty"`
	Volumes         map[string]struct{}       `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumeDriver    string                    `json:"VolumeDriver,omitempty" yaml:"VolumeDriver,omitempty"`
	WorkingDir      string                    `json:"WorkingDir,omitempty" yaml:"WorkingDir,omitempty"`
	Entrypoint      []string                  `json:"Entrypoint" yaml:"Entrypoint"`
	NetworkDisabled bool                      `json:"NetworkDisabled,omitempty" yaml:"NetworkDisabled,omitempty"`
	MacAddress      string                    `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	OnBuild         []string                  `json:"OnBuild,omitempty" yaml:"OnBuild,omitempty"`
	Labels          map[string]string         `json:"Labels,omitempty" yaml:"Labels,omitempty"`

	// FIXME: The following fields have been removed since API v1.18
	Memory          int64                     `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap      int64                     `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	CpuShares       int64                     `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	Cpuset          string                    `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	PortSpecs       []string                  `json:"PortSpecs,omitempty" yaml:"PortSpecs,omitempty"`

	// This is used only by the create command
	HostConfig HostConfig
}

type HostConfig struct {
	Binds           []string                  `json:"Binds,omitempty" yaml:"Binds,omitempty"`
	ContainerIDFile string                    `json:"ContainerIDFile,omitempty" yaml:"ContainerIDFile,omitempty"`
	LxcConf         []map[string]string       `json:"LxcConf,omitempty" yaml:"LxcConf,omitempty"`
	Memory          int64                     `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap      int64                     `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	CpuShares       int64                     `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CpuPeriod       int64                     `json:"CpuPeriod,omitempty" yaml:"CpuPeriod,omitempty"`
	CpusetCpus      string                    `json:"CpusetCpus,omitempty" yaml:"CpusetCpus,omitempty"`
	CpusetMems      string                    `json:"CpusetMems,omitempty" yaml:"CpusetMems,omitempty"`
	CpuQuota        int64                     `json:"CpuQuota,omitempty" yaml:"CpuQuota,omitempty"`
	BlkioWeight     int64                     `json:"BlkioWeight,omitempty" yaml:"BlkioWeight,omitempty"`
	OomKillDisable  bool                      `json:"OomKillDisable,omitempty" yaml:"OomKillDisable,omitempty"`
	Privileged      bool                      `json:"Privileged,omitempty" yaml:"Privileged,omitempty"`
	PortBindings    map[string][]PortBinding  `json:"PortBindings,omitempty" yaml:"PortBindings,omitempty"`
	Links           []string                  `json:"Links,omitempty" yaml:"Links,omitempty"`
	PublishAllPorts bool                      `json:"PublishAllPorts,omitempty" yaml:"PublishAllPorts,omitempty"`
	Dns             []string                  `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.10 and above only
	DnsSearch       []string                  `json:"DnsSearch,omitempty" yaml:"DnsSearch,omitempty"`
	ExtraHosts      []string                  `json:"ExtraHosts,omitempty" yaml:"ExtraHosts,omitempty"`
	VolumesFrom     []string                  `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
	Devices         []DeviceMapping           `json:"Devices,omitempty" yaml:"Devices,omitempty"`
	NetworkMode     string                    `json:"NetworkMode,omitempty" yaml:"NetworkMode,omitempty"`
	IpcMode         string                    `json:"IpcMode,omitempty" yaml:"IpcMode,omitempty"`
	PidMode         string                    `json:"PidMode,omitempty" yaml:"PidMode,omitempty"`
	UTSMode         string                    `json:"UTSMode,omitempty" yaml:"UTSMode,omitempty"`
	CapAdd          []string                  `json:"CapAdd,omitempty" yaml:"CapAdd,omitempty"`
	CapDrop         []string                  `json:"CapDrop,omitempty" yaml:"CapDrop,omitempty"`
	RestartPolicy   RestartPolicy             `json:"RestartPolicy,omitempty" yaml:"RestartPolicy,omitempty"`
	SecurityOpt     []string                  `json:"SecurityOpt,omitempty" yaml:"SecurityOpt,omitempty"`
	ReadonlyRootfs  bool                      `json:"ReadonlyRootfs,omitempty" yaml:"ReadonlyRootfs,omitempty"`
	Ulimits         []Ulimit                  `json:"Ulimits,omitempty" yaml:"Ulimits,omitempty"`
	LogConfig       LogConfig                 `json:"LogConfig,omitempty" yaml:"LogConfig,omitempty"`
	CgroupParent    string                    `json:"CgroupParent,omitempty" yaml:"CgroupParent,omitempty"`
}

type DeviceMapping struct {
	PathOnHost        string                  `json:"PathOnHost" yaml:"PathOnHost"`
	PathInContainer   string                  `json:"PathInContainer" yaml:"PathInContainer"`
	CgroupPermissions string                  `json:"CgroupPermissions" yaml:"CgroupPermissions"`
}

type ExecConfig struct {
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Tty          bool
	Cmd          []string
	Container    string
	Detach       bool
}

type LogOptions struct {
	Follow     bool
	Stdout     bool
	Stderr     bool
	Timestamps bool
	Tail       int64
}

type MonitorEventsFilters struct {
	Event     string `json:",omitempty"`
	Image     string `json:",omitempty"`
	Container string `json:",omitempty"`
}

type MonitorEventsOptions struct {
	Since   int
	Until   int
	Filters *MonitorEventsFilters `json:",omitempty"`
}

type RestartPolicy struct {
	Name              string                  `json:"Name,omitempty" yaml:"Name,omitempty"`
	MaximumRetryCount int64                   `json:"MaximumRetryCount,omitempty" yaml:"MaximumRetryCount,omitempty"`
}

type PortBinding struct {
	HostIp   string                           `json:"HostIp,omitempty" yaml:"HostIp,omitempty"`
	HostPort string                           `json:"HostPort,omitempty" yaml:"HostPort,omitempty"`
}

type State struct {
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        int
	ExitCode   int
	Error      string // contains last known error when starting the container
	StartedAt  time.Time
	FinishedAt time.Time
	Ghost      bool
}

// String returns a human-readable description of the state
// Stoken from docker/docker/daemon/state.go
func (s *State) String() string {
	if s.Running {
		if s.Paused {
			return fmt.Sprintf("Up %s (Paused)", units.HumanDuration(time.Now().UTC().Sub(s.StartedAt)))
		}
		if s.Restarting {
			return fmt.Sprintf("Restarting (%d) %s ago", s.ExitCode, units.HumanDuration(time.Now().UTC().Sub(s.FinishedAt)))
		}

		return fmt.Sprintf("Up %s", units.HumanDuration(time.Now().UTC().Sub(s.StartedAt)))
	}

	if s.Dead {
		return "Dead"
	}

	if s.FinishedAt.IsZero() {
		return ""
	}

	return fmt.Sprintf("Exited (%d) %s ago", s.ExitCode, units.HumanDuration(time.Now().UTC().Sub(s.FinishedAt)))
}

// StateString returns a single string to describe state
// Stoken from docker/docker/daemon/state.go
func (s *State) StateString() string {
	if s.Running {
		if s.Paused {
			return "paused"
		}
		if s.Restarting {
			return "restarting"
		}
		return "running"
	}

	if s.Dead {
		return "dead"
	}

	return "exited"
}

type ImageInfo struct {
	Architecture    string
	Author          string
	Comment         string
	Config          *ContainerConfig
	Container       string
	ContainerConfig *ContainerConfig
	Created         time.Time
	DockerVersion   string
	Id              string
	Os              string
	Parent          string
	Size            int64
	VirtualSize     int64
}

type ContainerInfo struct {
	Id              string
	Created         string
	Path            string
	Name            string
	Args            []string
	ExecIDs         []string
	Config          *ContainerConfig
	State           *State
	Image           string
	NetworkSettings struct {
		IPAddress   string `json:"IpAddress"`
		IPPrefixLen int    `json:"IpPrefixLen"`
		Gateway     string
		Bridge      string
		Ports       map[string][]PortBinding
	}
	SysInitPath    string
	ResolvConfPath string
	Volumes        map[string]string
	HostConfig     *HostConfig
}

type ContainerChanges struct {
	Path string
	Kind int
}

type Port struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

type Container struct {
	Id         string
	Names      []string
	Image      string
	Command    string
	Created    int64
	Status     string
	Ports      []Port
	SizeRw     int64
	SizeRootFs int64
	Labels     map[string]string
}

type Event struct {
	Id     string
	Status string
	From   string
	Time   int64
}

type Version struct {
	ApiVersion    string
	Arch          string
	GitCommit     string
	GoVersion     string
	KernelVersion string
	Os            string
	Version       string
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

// Info is the struct returned by /info
// The API is currently in flux, so Debug, MemoryLimit, SwapLimit, and
// IPv4Forwarding are interfaces because in docker 1.6.1 they are 0 or 1 but in
// master they are bools.
type Info struct {
	ID                 string
	Containers         int64
	Driver             string
	DriverStatus       [][]string
	ExecutionDriver    string
	Images             int64
	KernelVersion      string
	OperatingSystem    string
	NCPU               int64
	MemTotal           int64
	Name               string
	Labels             []string
	Debug              interface{}
	NFd                int64
	NGoroutines        int64
	SystemTime         string
	NEventsListener    int64
	InitPath           string
	InitSha1           string
	IndexServerAddress string
	MemoryLimit        interface{}
	SwapLimit          interface{}
	IPv4Forwarding     interface{}
	BridgeNfIptables   bool
	BridgeNfIp6tables  bool
	DockerRootDir      string
	HttpProxy          string
	HttpsProxy         string
	NoProxy            string
}

type ImageDelete struct {
	Deleted  string
	Untagged string
}

type EventOrError struct {
	Event
	Error error
}

type decodingResult struct {
	result interface{}
	err    error
}

// The following are types for the API stats endpoint
type ThrottlingData struct {
	// Number of periods with throttling active
	Periods uint64 `json:"periods"`
	// Number of periods when the container hit its throttling limit.
	ThrottledPeriods uint64 `json:"throttled_periods"`
	// Aggregate time the container was throttled for in nanoseconds.
	ThrottledTime uint64 `json:"throttled_time"`
}

type CpuUsage struct {
	// Total CPU time consumed.
	// Units: nanoseconds.
	TotalUsage uint64 `json:"total_usage"`
	// Total CPU time consumed per core.
	// Units: nanoseconds.
	PercpuUsage []uint64 `json:"percpu_usage"`
	// Time spent by tasks of the cgroup in kernel mode.
	// Units: nanoseconds.
	UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
	// Time spent by tasks of the cgroup in user mode.
	// Units: nanoseconds.
	UsageInUsermode uint64 `json:"usage_in_usermode"`
}

type CpuStats struct {
	CpuUsage       CpuUsage       `json:"cpu_usage"`
	SystemUsage    uint64         `json:"system_cpu_usage"`
	ThrottlingData ThrottlingData `json:"throttling_data,omitempty"`
}

type NetworkStats struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}

type MemoryStats struct {
	Usage    uint64            `json:"usage"`
	MaxUsage uint64            `json:"max_usage"`
	Stats    map[string]uint64 `json:"stats"`
	Failcnt  uint64            `json:"failcnt"`
	Limit    uint64            `json:"limit"`
}

type BlkioStatEntry struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Op    string `json:"op"`
	Value uint64 `json:"value"`
}

type BlkioStats struct {
	// number of bytes tranferred to and from the block device
	IoServiceBytesRecursive []BlkioStatEntry `json:"io_service_bytes_recursive"`
	IoServicedRecursive     []BlkioStatEntry `json:"io_serviced_recursive"`
	IoQueuedRecursive       []BlkioStatEntry `json:"io_queue_recursive"`
	IoServiceTimeRecursive  []BlkioStatEntry `json:"io_service_time_recursive"`
	IoWaitTimeRecursive     []BlkioStatEntry `json:"io_wait_time_recursive"`
	IoMergedRecursive       []BlkioStatEntry `json:"io_merged_recursive"`
	IoTimeRecursive         []BlkioStatEntry `json:"io_time_recursive"`
	SectorsRecursive        []BlkioStatEntry `json:"sectors_recursive"`
}

type Stats struct {
	Read         time.Time    `json:"read"`
	NetworkStats NetworkStats `json:"network,omitempty"`
	CpuStats     CpuStats     `json:"cpu_stats,omitempty"`
	MemoryStats  MemoryStats  `json:"memory_stats,omitempty"`
	BlkioStats   BlkioStats   `json:"blkio_stats,omitempty"`
}

type Ulimit struct {
	Name string                                 `json:"name" yaml:"name"`
	Soft uint64                                 `json:"soft" yaml:"soft"`
	Hard uint64                                 `json:"hard" yaml:"hard"`
}

type LogConfig struct {
	Type   string                               `json:"type" yaml:"type"`
	Config map[string]string                    `json:"config" yaml:"config"`
}

type BuildImage struct {
	Config         *ConfigFile
	DockerfileName string
	Context        io.Reader
	RemoteURL      string
	RepoName       string
	SuppressOutput bool
	NoCache        bool
	Remove         bool
	ForceRemove    bool
	Pull           bool
	Memory         int64
	MemorySwap     int64
	CpuShares      int64
	CpuPeriod      int64
	CpuQuota       int64
	CpuSetCpus     string
	CpuSetMems     string
	CgroupParent   string
}
