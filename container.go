package testcontainers

import (
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func dockerCmd() string {
	if runtime.GOOS == "windows" {
		return "docker.exe"
	}
	return "docker"
}

func DockerExists() bool {
	out, err := exec.Command(dockerCmd(), "version").Output()
	if err != nil {
		return false
	}
	cnt := strings.Count(string(out), "Version")
	return cnt > 0
}

func ContainerExists(name string) (bool, error) {
	out, err := exec.Command(dockerCmd(), "inspect", "--format=\"{{.Name}}\"", name).Output()
	if err != nil {
		if noSuchContainerErr(err) {
			return false, nil
		}
		return false, err
	}
	containerName := strings.Trim(string(out), "\n")
	containerName = strings.Trim(containerName, `"`)
	return "/"+name == containerName, nil
}

func DropContainerIfExists(containerName string) {
	for i := 0; i < 3; i++ {
		out, err := exec.Command(dockerCmd(), "container", "rm", "-f", "/"+containerName).Output()
		if noSuchContainerErr(err) {
			return
		}
		if len(out) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
}

func DropNetwork(networkName string) {
	_, _ = exec.Command(dockerCmd(), "network", "rm", networkName).Output()
}

func DropContainers(containerNames []string) {
	_, _ = exec.Command(dockerCmd(), "network", "-f", "prune").Output()
	for i := 0; i < len(containerNames); i++ {
		DropContainerIfExists(containerNames[i])
	}
}

func noSuchContainerErr(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		msg := string(exitErr.Stderr)
		if strings.Contains(msg, "Error: No such") {
			return true
		}
	}
	return false
}
