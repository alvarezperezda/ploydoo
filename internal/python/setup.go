package python

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// pythonVersionMap maps Odoo versions to their required Python versions.
var pythonVersionMap = map[string]string{
	"16.0": "3.10",
	"17.0": "3.10",
	"18.0": "3.12",
}

// pythonPatchVersionMap maps Python minor versions to a recommended full patch version.
var pythonPatchVersionMap = map[string]string{
	"3.10": "3.10.13",
	"3.12": "3.12.2",
}

// PythonVersionForOdoo returns the Python minor version needed for the given Odoo version.
func PythonVersionForOdoo(odooVersion string) string {
	if v, ok := pythonVersionMap[odooVersion]; ok {
		return v
	}
	return "3.12" // default for unknown versions
}

// PythonPatchVersion returns the full patch version (e.g. "3.10.13") for a minor version.
func PythonPatchVersion(minorVersion string) string {
	if v, ok := pythonPatchVersionMap[minorVersion]; ok {
		return v
	}
	return minorVersion
}

// CheckPythonAvailable checks if a specific Python version (e.g. "3.10") is available on the system.
func CheckPythonAvailable(version string) bool {
	// Try python3.X --version
	cmd := exec.Command(fmt.Sprintf("python%s", version), "--version")
	if err := cmd.Run(); err == nil {
		return true
	}

	// Try via pyenv shims
	home, err := os.UserHomeDir()
	if err == nil {
		shimPath := filepath.Join(home, ".pyenv", "shims", fmt.Sprintf("python%s", version))
		if _, err := os.Stat(shimPath); err == nil {
			cmd := exec.Command(shimPath, "--version")
			if err := cmd.Run(); err == nil {
				return true
			}
		}
	}

	return false
}

// CheckPyenvAvailable checks if pyenv is installed and available.
func CheckPyenvAvailable() bool {
	_, err := exec.LookPath("pyenv")
	return err == nil
}

// CheckPoetryAvailable checks if poetry is installed and available.
func CheckPoetryAvailable() bool {
	_, err := exec.LookPath("poetry")
	return err == nil
}

// InstallPythonWithPyenv installs the given Python version using pyenv.
func InstallPythonWithPyenv(version string) error {
	patchVersion := PythonPatchVersion(version)
	cmd := exec.Command("pyenv", "install", "--skip-existing", patchVersion)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// GeneratePyprojectToml parses the Odoo requirements.txt and generates a pyproject.toml.
func GeneratePyprojectToml(baseDir, pythonVersion string) error {
	reqPath := filepath.Join(baseDir, "odoo", "requirements.txt")
	deps, err := parseRequirements(reqPath, pythonVersion)
	if err != nil {
		return fmt.Errorf("failed to parse requirements.txt: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("[tool.poetry]\n")
	sb.WriteString("name = \"odoo-project\"\n")
	sb.WriteString("version = \"1.0.0\"\n")
	sb.WriteString("description = \"Odoo development environment\"\n")
	sb.WriteString("authors = [\"developer\"]\n")
	sb.WriteString("package-mode = false\n")
	sb.WriteString("\n")
	sb.WriteString("[tool.poetry.dependencies]\n")
	sb.WriteString(fmt.Sprintf("python = \"^%s\"\n", pythonVersion))

	for _, dep := range deps {
		sb.WriteString(dep + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString("[build-system]\n")
	sb.WriteString("requires = [\"poetry-core\"]\n")
	sb.WriteString("build-backend = \"poetry.core.masonry.api\"\n")

	pyprojectPath := filepath.Join(baseDir, "pyproject.toml")
	return os.WriteFile(pyprojectPath, []byte(sb.String()), 0644)
}

// parseRequirements reads a requirements.txt and returns poetry-compatible dependency lines.
// It filters by platform (excludes win32-only) and python version, and deduplicates.
func parseRequirements(path, pythonVersion string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	seen := make(map[string]string) // lowercase name -> toml line
	var order []string              // preserve insertion order
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		name, dep := convertRequirement(line, pythonVersion)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; !exists {
			order = append(order, key)
		}
		seen[key] = dep
	}

	var deps []string
	for _, key := range order {
		deps = append(deps, seen[key])
	}
	return deps, scanner.Err()
}

// convertRequirement converts a pip requirement line to a poetry TOML line.
// Filters out Windows-only packages and packages that don't match the target Python version.
// Returns ("", "") if the line should be skipped.
func convertRequirement(line, pythonVersion string) (string, string) {
	// Remove inline comments
	if idx := strings.Index(line, " #"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}

	// Extract and evaluate environment markers
	markers := ""
	if idx := strings.Index(line, ";"); idx >= 0 {
		markers = strings.TrimSpace(line[idx+1:])
		line = strings.TrimSpace(line[:idx])
	}

	if markers != "" && !evaluateMarkers(markers, pythonVersion) {
		return "", ""
	}

	name := line
	version := ""

	// Find first version specifier
	for _, sep := range []string{">=", "<=", "!=", "==", "~=", ">", "<"} {
		if idx := strings.Index(line, sep); idx >= 0 {
			name = strings.TrimSpace(line[:idx])
			version = strings.TrimSpace(line[idx:])
			break
		}
	}

	// Remove extras from name (e.g. "package[extra]")
	if idx := strings.Index(name, "["); idx >= 0 {
		name = name[:idx]
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return "", ""
	}

	if version == "" {
		return name, fmt.Sprintf("%s = \"*\"", name)
	}
	return name, fmt.Sprintf("%s = \"%s\"", name, version)
}

// evaluateMarkers checks if environment markers match our target platform (not win32)
// and target Python version. This is a simplified evaluator that handles the common cases
// in Odoo's requirements.txt.
func evaluateMarkers(markers, pythonVersion string) bool {
	// Skip packages that are Windows-only
	if strings.Contains(markers, "sys_platform == 'win32'") ||
		strings.Contains(markers, "sys_platform==\"win32\"") ||
		strings.Contains(markers, `sys_platform == "win32"`) {
		// Check if it's "== win32" (Windows only) vs "!= win32" (not Windows)
		if !strings.Contains(markers, "!=") {
			return false
		}
	}

	// Evaluate python_version conditions
	// Split by "and" to check all conditions
	parts := strings.Split(markers, " and ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "python_version") {
			if !evalPythonVersionCondition(part, pythonVersion) {
				return false
			}
		}
		// sys_platform != 'win32' is fine on macOS/Linux — pass through
	}
	return true
}

// evalPythonVersionCondition evaluates a single python_version condition.
// e.g. "python_version >= '3.12'" with pythonVersion "3.12" -> true
func evalPythonVersionCondition(cond, pythonVersion string) bool {
	// Extract operator and value
	cond = strings.ReplaceAll(cond, "python_version", "")
	cond = strings.TrimSpace(cond)

	var op, val string
	for _, o := range []string{">=", "<=", "!=", "==", ">", "<"} {
		if strings.HasPrefix(cond, o) {
			op = o
			val = strings.Trim(strings.TrimSpace(cond[len(o):]), "'\"")
			break
		}
	}
	if op == "" || val == "" {
		return true // can't parse, include it
	}

	cmp := compareVersions(pythonVersion, val)
	switch op {
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	}
	return true
}

// compareVersions compares two version strings (e.g. "3.10" vs "3.12").
// Returns -1, 0, or 1.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aNum, bNum int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &aNum)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bNum)
		}
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
	}
	return 0
}

// PoetrySetup runs poetry env use and poetry install in the given directory.
func PoetrySetup(baseDir, pythonVersion string) error {
	pythonBin := fmt.Sprintf("python%s", pythonVersion)

	// Try to find the python binary via pyenv shims if not in PATH
	if _, err := exec.LookPath(pythonBin); err != nil {
		home, homeErr := os.UserHomeDir()
		if homeErr == nil {
			shimPath := filepath.Join(home, ".pyenv", "shims", pythonBin)
			if _, err := os.Stat(shimPath); err == nil {
				pythonBin = shimPath
			}
		}
	}

	// poetry env use
	cmd := exec.Command("poetry", "env", "use", pythonBin)
	cmd.Dir = baseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("poetry env use failed: %w: %s", err, string(output))
	}

	// poetry install
	cmd = exec.Command("poetry", "install")
	cmd.Dir = baseDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("poetry install failed: %w: %s", err, string(output))
	}

	return nil
}
