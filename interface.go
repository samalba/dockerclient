package dockerclient

import (
	"io"
)

type Callback func(*Event, chan error, ...interface{})

type StatCallback func(string, *Stats, chan error, ...interface{})

type Client interface {
	Info() (*Info, error)
	ListContainers(all, size bool, filters string) ([]Container, error)
	InspectContainer(id string) (*ContainerInfo, error)
	CreateContainer(config *ContainerConfig, name string) (string, error)
	ContainerLogs(id string, options *LogOptions) (io.ReadCloser, error)
	ContainerChanges(id string) ([]*ContainerChanges, error)
	Exec(config *ExecConfig) (string, error)
	StartContainer(id string, config *HostConfig) error
	StopContainer(id string, timeout int) error
	RestartContainer(id string, timeout int) error
	KillContainer(id, signal string) error
	// MonitorEvents returns an EventOrError channel and a close channel. If
	// an error is ever sent, then no more events will be sent. Users must
	// always close the close channel when they are done reading events,
	// even if an error was sent.
	MonitorEvents(options *MonitorEventsOptions) (<-chan EventOrError, chan<- struct{}, error)
	StartMonitorEvents(cb Callback, ec chan error, args ...interface{})
	StopAllMonitorEvents()
	StartMonitorStats(id string, cb StatCallback, ec chan error, args ...interface{})
	StopAllMonitorStats()
	TagImage(nameOrID string, repo string, tag string, force bool) error
	Version() (*Version, error)
	PullImage(name string, auth *AuthConfig) error
	LoadImage(reader io.Reader) error
	RemoveContainer(id string, force, volumes bool) error
	ListImages() ([]*Image, error)
	RemoveImage(name string) ([]*ImageDelete, error)
	PauseContainer(name string) error
	UnpauseContainer(name string) error
	RenameContainer(oldName string, newName string) error
	ImportImage(source string, repository string, tag string, tar io.Reader) (io.ReadCloser, error)
}
