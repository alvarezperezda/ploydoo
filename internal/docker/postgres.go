package docker

import (
	"fmt"
	"os/exec"
)

// StartPostgres runs a detached PostgreSQL container with the given configuration.
func StartPostgres(pgVersion, odooVersion, dbUser, dbPassword, dbName string) error {
	containerName := fmt.Sprintf("odoo-postgres-%s", odooVersion)
	image := fmt.Sprintf("postgres:%s", pgVersion)

	args := []string{
		"run", "-d", "--rm",
		"--name", containerName,
		"-e", fmt.Sprintf("POSTGRES_USER=%s", dbUser),
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
		"-e", fmt.Sprintf("POSTGRES_DB=%s", dbName),
		"-p", "5432:5432",
		image,
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
