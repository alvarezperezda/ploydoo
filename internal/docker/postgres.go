package docker

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"time"
)

// IsPortInUse checks if a given port is already in use.
func IsPortInUse(port string) bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+port, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsDockerRunning checks if the Docker daemon is running.
func IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

// StartDocker attempts to start the Docker daemon.
func StartDocker() error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("open", "-a", "Docker")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start Docker Desktop: %w", err)
		}
	case "linux":
		cmd := exec.Command("sudo", "systemctl", "start", "docker")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to start docker service: %w: %s", err, string(output))
		}
	default:
		return fmt.Errorf("cannot auto-start Docker on %s", runtime.GOOS)
	}

	// Wait for Docker to be ready (up to 30 seconds)
	for i := 0; i < 30; i++ {
		if IsDockerRunning() {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Docker did not start within 30 seconds")
}

// StartPostgres runs a detached PostgreSQL container with the given configuration.
func StartPostgres(pgVersion, odooVersion, dbUser, dbPassword, dbName, port string) error {
	// Ensure Docker is running
	if !IsDockerRunning() {
		if err := StartDocker(); err != nil {
			return fmt.Errorf("Docker is not running and could not be started: %w", err)
		}
	}

	containerName := fmt.Sprintf("odoo-postgres-%s", odooVersion)
	image := fmt.Sprintf("postgres:%s", pgVersion)

	args := []string{
		"run", "-d", "--rm",
		"--name", containerName,
		"-e", fmt.Sprintf("POSTGRES_USER=%s", dbUser),
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
		"-e", fmt.Sprintf("POSTGRES_DB=%s", dbName),
		"-p", fmt.Sprintf("%s:5432", port),
		image,
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
