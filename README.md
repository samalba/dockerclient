Docker client library in Go
===========================

This library supports few API calls but it will get extended over time.

Example:

```go
package main

import (
	"github.com/samalba/dockerclient"
	"log"
)

func main() {
	docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock")
	// Get only running containers
	containers, err := docker.ListContainers(false)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range *containers {
		log.Println(c.Id, c.Names)
	}
	// Inspect the first container returned
	info, _ := docker.InspectContainer((*containers)[0].Id)
	log.Println(info)
}
```
