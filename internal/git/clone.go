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
	dest := filepath.Join(baseDir, "addons", "OCA", module)
	return runClone(url, version, dest, 1)
}

// CloneCustomAddons clones a custom addons repository with the given branch.
func CloneCustomAddons(baseDir, repoURL, branch string) error {
	dest := filepath.Join(baseDir, "addons", "custom_addons")
	return runClone(repoURL, branch, dest, 1)
}

// ListRemoteBranches returns the list of remote branches for the given repository URL.
func ListRemoteBranches(repoURL string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", repoURL)
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
