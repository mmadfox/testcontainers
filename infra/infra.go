package infra

import (
	"os/exec"
	"strings"
	"time"
)

func ExistsContainer(name string) (bool, error) {
	out, err := exec.Command("docker", "inspect", "--format=\"{{.Name}}\"", name).Output()
	if err != nil {
		return false, err
	}
	containerName := strings.Trim(string(out), "\n")
	containerName = strings.Trim(containerName, `"`)
	return "/"+name == containerName, nil
}

func DropContainerIfExists(containerName string) {
	for i := 0; i < 3; i++ {
		out, _ := exec.Command("docker", "container", "rm", "-f", "/"+containerName).Output()
		if len(out) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
}
