package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"odoo-cli/internal/config"
	"odoo-cli/internal/docker"
	gitops "odoo-cli/internal/git"
	"odoo-cli/internal/python"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// step represents which screen the TUI is on.
type step int

const (
	stepPath step = iota
	stepVersion
	stepModules
	stepCustomAddons
	stepCustomURL
	stepCustomBranch
	stepPostgres
	stepDBConfig
	stepPython
	stepCloning
	stepDone
)

// Model is the main Bubbletea model for the CLI.
type Model struct {
	step step
	Err  error

	// Terminal size
	termWidth  int
	termHeight int

	// Installation path
	installPath string
	pathErr     string

	// Version selection
	versionCursor int
	version       string

	// Module selection
	moduleCursor   int
	moduleSelected map[int]bool
	moduleOffset   int // scroll offset for the module list viewport

	// PostgreSQL version selection
	pgCursor  int
	pgVersion string

	// Database configuration
	dbFields      [dbFieldCount]string
	dbActiveField int
	dbErr         string

	// Custom addons
	customAddons   bool
	customURL      string
	customURLErr   string
	customBranches []string
	customCursor   int
	customOffset   int
	customBranch   string
	customLoading  bool
	customBranchErr string

	// Python environment
	pythonVersion   string
	pythonAvailable bool
	pyenvAvailable  bool
	poetryAvailable bool
	installPython   bool
	installPoetry   bool

	// Cloning progress
	spinner        spinner.Model
	cloneResults   []CloneStatus
	currentTask    string
	cloneQueue     []cloneJob
	cloneIndex     int
	progressOffset int
}

type cloneJob struct {
	name string
	fn   func() error
}

// Messages
type cloneResultMsg struct {
	name    string
	err     error
}

type allDoneMsg struct{}

type customBranchesMsg struct {
	branches []string
	err      error
}

// NewModel creates a new TUI model.
func NewModel() Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7"))
	return Model{
		step:           stepPath,
		moduleSelected: make(map[int]bool),
		spinner:        sp,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.termWidth = msg.Width
		m.termHeight = msg.Height
	}
	switch m.step {
	case stepPath:
		return m.updatePath(msg)
	case stepVersion:
		return m.updateVersion(msg)
	case stepModules:
		return m.updateModules(msg)
	case stepCustomAddons:
		return m.updateCustomAddons(msg)
	case stepCustomURL:
		return m.updateCustomURL(msg)
	case stepCustomBranch:
		return m.updateCustomBranch(msg)
	case stepPostgres:
		return m.updatePostgres(msg)
	case stepDBConfig:
		return m.updateDBConfig(msg)
	case stepPython:
		return m.updatePython(msg)
	case stepCloning:
		return m.updateCloning(msg)
	case stepDone:
		return m.updateDone(msg)
	}
	return m, nil
}

func (m Model) View() string {
	switch m.step {
	case stepPath:
		return pathView(m.installPath, m.pathErr)
	case stepVersion:
		return versionView(m.versionCursor)
	case stepModules:
		return modulesView(m.moduleCursor, m.moduleSelected, m.moduleOffset, m.termHeight)
	case stepCustomAddons:
		return customAddonsView()
	case stepCustomURL:
		return customURLView(m.customURL, m.customURLErr)
	case stepCustomBranch:
		return customBranchView(m.customBranches, m.customCursor, m.customLoading, m.customBranchErr, m.customOffset, m.termHeight)
	case stepPostgres:
		return postgresView(m.pgCursor)
	case stepDBConfig:
		return dbConfigView(m.dbFields, m.dbActiveField, m.dbErr)
	case stepPython:
		return pythonView(m.pythonVersion, m.pythonAvailable, m.pyenvAvailable, m.poetryAvailable)
	case stepCloning:
		return progressView(m.cloneResults, m.currentTask, false, m.spinner.View(), m.progressOffset, m.termHeight)
	case stepDone:
		return progressView(m.cloneResults, "", true, "", m.progressOffset, m.termHeight)
	}
	return ""
}

// --- Path step ---

func (m Model) updatePath(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyBackspace:
			if len(m.installPath) > 0 {
				m.installPath = m.installPath[:len(m.installPath)-1]
			}
			m.pathErr = ""
		case tea.KeyEnter:
			path := expandHome(m.installPath)
			if path == "" {
				m.pathErr = "Path cannot be empty"
				return m, nil
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				m.pathErr = "Invalid path"
				return m, nil
			}
			// Create directory if it doesn't exist
			if err := os.MkdirAll(absPath, 0755); err != nil {
				m.pathErr = fmt.Sprintf("Cannot create directory: %v", err)
				return m, nil
			}
			m.installPath = absPath
			m.pathErr = ""
			m.step = stepVersion
		case tea.KeyRunes:
			m.installPath += string(msg.Runes)
			m.pathErr = ""
		case tea.KeySpace:
			m.installPath += " "
			m.pathErr = ""
		}
	}
	return m, nil
}

// --- Version step ---

func (m Model) updateVersion(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.versionCursor > 0 {
				m.versionCursor--
			}
		case "down", "j":
			if m.versionCursor < len(odooVersions)-1 {
				m.versionCursor++
			}
		case "enter":
			m.version = odooVersions[m.versionCursor]
			m.step = stepModules
		}
	}
	return m, nil
}

// --- Modules step ---

func (m Model) updateModules(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.moduleCursor > 0 {
				m.moduleCursor--
			}
		case "down", "j":
			if m.moduleCursor < len(ocaModules)-1 {
				m.moduleCursor++
			}
		case " ":
			m.moduleSelected[m.moduleCursor] = !m.moduleSelected[m.moduleCursor]
		case "a":
			for i := range ocaModules {
				m.moduleSelected[i] = true
			}
		case "n":
			for i := range ocaModules {
				m.moduleSelected[i] = false
			}
		case "enter":
			m.step = stepCustomAddons
		}
	}

	// Adjust scroll offset to keep cursor visible
	visible := modulesVisibleCount(m.termHeight)
	if m.moduleCursor < m.moduleOffset {
		m.moduleOffset = m.moduleCursor
	}
	if m.moduleCursor >= m.moduleOffset+visible {
		m.moduleOffset = m.moduleCursor - visible + 1
	}

	return m, nil
}

// --- Custom Addons step ---

func (m Model) updateCustomAddons(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "y":
			m.customAddons = true
			m.step = stepCustomURL
		case "n":
			m.customAddons = false
			m.step = stepPostgres
		}
	}
	return m, nil
}

// --- Custom URL step ---

func (m Model) updateCustomURL(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyBackspace:
			if len(m.customURL) > 0 {
				m.customURL = m.customURL[:len(m.customURL)-1]
			}
			m.customURLErr = ""
		case tea.KeyEnter:
			if m.customURL == "" {
				m.customURLErr = "URL cannot be empty"
				return m, nil
			}
			m.customURLErr = ""
			m.customLoading = true
			m.step = stepCustomBranch
			repoURL := m.customURL
			return m, func() tea.Msg {
				branches, err := gitops.ListRemoteBranches(repoURL)
				return customBranchesMsg{branches: branches, err: err}
			}
		case tea.KeyRunes:
			m.customURL += string(msg.Runes)
			m.customURLErr = ""
		case tea.KeySpace:
			m.customURL += " "
			m.customURLErr = ""
		}
	}
	return m, nil
}

// --- Custom Branch step ---

func (m Model) updateCustomBranch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case customBranchesMsg:
		m.customLoading = false
		if msg.err != nil {
			m.customBranchErr = fmt.Sprintf("Failed to fetch branches: %v", msg.err)
		} else {
			m.customBranches = msg.branches
		}
		return m, nil

	case tea.KeyMsg:
		if m.customLoading {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.customCursor > 0 {
				m.customCursor--
			}
		case "down", "j":
			if m.customCursor < len(m.customBranches)-1 {
				m.customCursor++
			}
		case "enter":
			if len(m.customBranches) > 0 {
				m.customBranch = m.customBranches[m.customCursor]
			}
			m.step = stepPostgres
		}
	}

	// Adjust scroll offset
	if len(m.customBranches) > 0 {
		visible := customBranchVisibleCount(m.termHeight)
		if m.customCursor < m.customOffset {
			m.customOffset = m.customCursor
		}
		if m.customCursor >= m.customOffset+visible {
			m.customOffset = m.customCursor - visible + 1
		}
	}

	return m, nil
}

// --- Postgres step ---

func (m Model) updatePostgres(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.pgCursor > 0 {
				m.pgCursor--
			}
		case "down", "j":
			if m.pgCursor < len(pgVersions)-1 {
				m.pgCursor++
			}
		case "enter":
			m.pgVersion = pgVersions[m.pgCursor]
			// Set defaults for DB config
			m.dbFields[dbFieldUser] = "odoo"
			m.dbFields[dbFieldPassword] = "odoo"
			m.dbFields[dbFieldName] = fmt.Sprintf("odoo-%s", m.version)
			// Detect if port 5432 is in use
			if docker.IsPortInUse("5432") {
				m.dbFields[dbFieldPort] = "5433"
				m.dbErr = "Port 5432 is in use — defaulting to 5433"
			} else {
				m.dbFields[dbFieldPort] = "5432"
			}
			m.step = stepDBConfig
		}
	}
	return m, nil
}

// --- DB Config step ---

func (m Model) updateDBConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if m.dbActiveField > 0 {
				m.dbActiveField--
			}
			m.dbErr = ""
		case tea.KeyDown:
			if m.dbActiveField < dbFieldCount-1 {
				m.dbActiveField++
			}
			m.dbErr = ""
		case tea.KeyTab:
			m.dbActiveField = (m.dbActiveField + 1) % dbFieldCount
			m.dbErr = ""
		case tea.KeyBackspace:
			f := m.dbFields[m.dbActiveField]
			if len(f) > 0 {
				m.dbFields[m.dbActiveField] = f[:len(f)-1]
			}
			m.dbErr = ""
		case tea.KeyEnter:
			if m.dbFields[dbFieldUser] == "" {
				m.dbErr = "User cannot be empty"
				return m, nil
			}
			if m.dbFields[dbFieldPassword] == "" {
				m.dbErr = "Password cannot be empty"
				return m, nil
			}
			if m.dbFields[dbFieldName] == "" {
				m.dbErr = "Database name cannot be empty"
				return m, nil
			}
			if m.dbFields[dbFieldPort] == "" {
				m.dbErr = "Port cannot be empty"
				return m, nil
			}
			// Determine Python version and check availability
			m.pythonVersion = python.PythonVersionForOdoo(m.version)
			m.pythonAvailable = python.CheckPythonAvailable(m.pythonVersion)
			m.pyenvAvailable = python.CheckPyenvAvailable()
			m.poetryAvailable = python.CheckPoetryAvailable()
			m.step = stepPython
		case tea.KeyRunes:
			key := string(msg.Runes)
			if key == "q" && m.dbFields[m.dbActiveField] == "" {
				return m, tea.Quit
			}
			m.dbFields[m.dbActiveField] += key
			m.dbErr = ""
		case tea.KeySpace:
			m.dbFields[m.dbActiveField] += " "
			m.dbErr = ""
		}
	}
	return m, nil
}

// --- Python step ---

func (m Model) updatePython(msg tea.Msg) (tea.Model, tea.Cmd) {
	needsInstall := (!m.pythonAvailable && m.pyenvAvailable) || !m.poetryAvailable

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !needsInstall {
				m.step = stepCloning
				m.buildCloneQueue()
				return m, tea.Batch(m.spinner.Tick, m.runNextClone())
			}
		case "y":
			if needsInstall {
				if !m.pythonAvailable && m.pyenvAvailable {
					m.installPython = true
				}
				if !m.poetryAvailable {
					m.installPoetry = true
				}
				m.step = stepCloning
				m.buildCloneQueue()
				return m, tea.Batch(m.spinner.Tick, m.runNextClone())
			}
		case "n":
			if needsInstall {
				m.step = stepCloning
				m.buildCloneQueue()
				return m, tea.Batch(m.spinner.Tick, m.runNextClone())
			}
		}
	}
	return m, nil
}

// --- Cloning step ---

func (m *Model) buildCloneQueue() {
	baseDir := m.installPath

	m.cloneQueue = append(m.cloneQueue, cloneJob{
		name: "odoo",
		fn: func() error {
			return gitops.CloneOdoo(baseDir, m.version)
		},
	})

	for i, mod := range ocaModules {
		if m.moduleSelected[i] {
			mod := mod // capture
			m.cloneQueue = append(m.cloneQueue, cloneJob{
				name: mod,
				fn: func() error {
					return gitops.CloneOCAModule(baseDir, mod, m.version)
				},
			})
		}
	}

	if m.customAddons && m.customURL != "" && m.customBranch != "" {
		customURL := m.customURL
		customBranch := m.customBranch
		m.cloneQueue = append(m.cloneQueue, cloneJob{
			name: "custom_addons",
			fn: func() error {
				return gitops.CloneCustomAddons(baseDir, customURL, customBranch)
			},
		})
	}

	m.cloneQueue = append(m.cloneQueue, cloneJob{
		name: "postgres-container",
		fn: func() error {
			return docker.StartPostgres(m.pgVersion, m.version, m.dbFields[dbFieldUser], m.dbFields[dbFieldPassword], m.dbFields[dbFieldName], m.dbFields[dbFieldPort])
		},
	})

	// Python environment jobs
	if m.installPython && m.pyenvAvailable {
		pyVer := m.pythonVersion
		m.cloneQueue = append(m.cloneQueue, cloneJob{
			name: "python-install",
			fn: func() error {
				return python.InstallPythonWithPyenv(pyVer)
			},
		})
	}

	if m.installPoetry {
		m.cloneQueue = append(m.cloneQueue, cloneJob{
			name: "poetry-install",
			fn: func() error {
				return python.InstallPoetry()
			},
		})
	}

	pyVer := m.pythonVersion
	installPath := m.installPath
	m.cloneQueue = append(m.cloneQueue, cloneJob{
		name: "pyproject-toml",
		fn: func() error {
			return python.GeneratePyprojectToml(installPath, pyVer)
		},
	})

	m.cloneQueue = append(m.cloneQueue, cloneJob{
		name: "poetry-setup",
		fn: func() error {
			return python.PoetrySetup(installPath, pyVer)
		},
	})

	m.cloneIndex = 0
	if len(m.cloneQueue) > 0 {
		m.currentTask = taskLabel(m.cloneQueue[0].name)
	}
}

func taskLabel(name string) string {
	switch name {
	case "postgres-container":
		return "Starting PostgreSQL container..."
	case "python-install":
		return "Installing Python with pyenv..."
	case "poetry-install":
		return "Installing Poetry..."
	case "pyproject-toml":
		return "Generating pyproject.toml..."
	case "poetry-setup":
		return "Setting up Poetry environment..."
	default:
		return "Cloning " + name + "..."
	}
}

func (m *Model) runNextClone() tea.Cmd {
	if m.cloneIndex >= len(m.cloneQueue) {
		return func() tea.Msg { return allDoneMsg{} }
	}
	job := m.cloneQueue[m.cloneIndex]
	return func() tea.Msg {
		err := job.fn()
		return cloneResultMsg{name: job.name, err: err}
	}
}

func (m Model) updateCloning(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case cloneResultMsg:
		m.cloneResults = append(m.cloneResults, CloneStatus{
			Name:    msg.name,
			Success: msg.err == nil,
			Err:     msg.err,
		})
		m.cloneIndex++

		// Auto-scroll to keep latest result visible
		visible := progressVisibleCount(m.termHeight)
		if len(m.cloneResults) > visible {
			m.progressOffset = len(m.cloneResults) - visible
		}

		if m.cloneIndex < len(m.cloneQueue) {
			m.currentTask = taskLabel(m.cloneQueue[m.cloneIndex].name)
			return m, m.runNextClone()
		}
		// All done, generate config
		return m, func() tea.Msg { return allDoneMsg{} }

	case allDoneMsg:
		var selectedMods []string
		for i, mod := range ocaModules {
			if m.moduleSelected[i] {
				selectedMods = append(selectedMods, mod)
			}
		}
		_ = config.GenerateOdooConf(m.installPath, selectedMods, m.customAddons, m.version, m.dbFields[dbFieldUser], m.dbFields[dbFieldPassword], m.dbFields[dbFieldName], m.dbFields[dbFieldPort])
		m.step = stepDone
		return m, nil
	}
	return m, nil
}

// --- Done step ---

func (m Model) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.progressOffset > 0 {
				m.progressOffset--
			}
		case "down", "j":
			visible := progressVisibleCount(m.termHeight)
			if m.progressOffset < len(m.cloneResults)-visible {
				m.progressOffset++
			}
		}
	}
	return m, nil
}
