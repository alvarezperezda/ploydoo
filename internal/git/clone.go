package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOdoo clones the main Odoo repository for the given version.
func CloneOdoo(baseDir, version string) error {
	dest := filepath.Join(baseDir, "odoo")
	return runClone("https://github.com/odoo/odoo.git", version, dest, 1)
}

// CloneOCAModule clones an OCA module repository for the given version.
func CloneOCAModule(baseDir, module, version string) error {
	url := fmt.Sprintf("https://github.com/OCA/%s.git", module)
	dest := filepath.Join(baseDir, "addons", module)
	return runClone(url, version, dest, 1)
}

// CloneAlventiaModules clones the alventia_modules repository with the given branch.
func CloneAlventiaModules(baseDir, branch string) error {
	dest := filepath.Join(baseDir, "addons", "alventia_modules")
	return runClone("git@github.com:daperez89/alventia_modules.git", branch, dest, 1)
}

// ListAlventiaBranches returns the list of remote branches for alventia_modules.
func ListAlventiaBranches() ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", "git@github.com:daperez89/alventia_modules.git")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(output))
	}

	var branches []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: <hash>\trefs/heads/<branch>
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			ref := parts[1]
			branch := strings.TrimPrefix(ref, "refs/heads/")
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

func runClone(url, branch, dest string, depth int) error {
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, "--depth", fmt.Sprintf("%d", depth))
	args = append(args, url, dest)

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
