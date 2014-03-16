## Docker client library in Go

This library supports few API calls but it will get extended over time.  

Docker is an open-source engine that automates the deployment of any application as a lightweight, 
portable, self-sufficient container that will run virtually anywhere.
For more information see [Learn what Docker is all about](https://www.docker.io/learn_more/)  

#### Example:

```go
package main

import (
	"github.com/samalba/dockerclient"
	"log"
	"os"
	"time"
)

var (
	_ = time.Now
)

// Callback used to listen to Docker's events
func eventCallback(event *dockerclient.Event, args ...interface{}) {
	log.Printf("Received event: %#v\n", *event)
}

func main() {
	// Init the client
	docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock")

	// List containers
	allList := false
	containers, err := docker.ListContainers(allList)
	if err != nil {
		log.Fatal("ListContainers: " + err.Error())
	} else {
		log.Println("ListContainers:")
		for _, c := range containers {
			log.Println(c.Id, c.Names)
		}
	}

	// Inspect the first container returned
	if len(containers) > 0 {
		id := containers[0].Id
		info, _ := docker.InspectContainer(id)
		log.Printf("InspectContainer: %+v\n", info)
	}

	// Create a container
	containerConfig := &dockerclient.ContainerConfig{
		Image:     "ubuntu:12.04",
		OpenStdin: true,
		Tty:       true,
	}
	containerConfig.Cmd = []string{"/bin/bash"}
	//containerConfig.Cmd = []string{"ls", "-alF"}
	//containerConfig.Cmd = []string{"echo"}

	containerId, err := docker.CreateContainer(containerConfig)
	if err != nil {
		log.Fatal("CreateContainer: " + err.Error())
	}

	// Start the container
	if containerId != "" {
		err = docker.StartContainer(containerId)
		if err != nil {
			log.Fatal("StartContainer: ", err.Error())
		}
	}

	// Attach the container
	if containerId != "" {
		att := dockerclient.Attach{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
		err = docker.AttachContainer(containerId, att)
		if err != nil {
			log.Fatal("AttachContainer: " + err.Error())
		}
	}

	// Stop the container (with 5 seconds timeout)
	if containerId != "" {
		docker.StopContainer(containerId, 5)
	}

	// Listen to the events
	// docker.StartMonitorEvents(eventCallback)
	// time.Sleep(3600 * time.Second)
}
```

### Contribution

Pull requests are welcome.

### License

Licensed under The Apache License  
For the full copyright and license information, please view the LICENSE file.