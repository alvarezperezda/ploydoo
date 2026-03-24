```
 ██████╗ ██╗      ██████╗ ██╗   ██╗██████╗  ██████╗  ██████╗
 ██╔══██╗██║     ██╔═══██╗╚██╗ ██╔╝██╔══██╗██╔═══██╗██╔═══██╗
 ██████╔╝██║     ██║   ██║ ╚████╔╝ ██║  ██║██║   ██║██║   ██║
 ██╔═══╝ ██║     ██║   ██║  ╚██╔╝  ██║  ██║██║   ██║██║   ██║
 ██║     ███████╗╚██████╔╝   ██║   ██████╔╝╚██████╔╝╚██████╔╝
 ╚═╝     ╚══════╝ ╚═════╝    ╚═╝   ╚═════╝  ╚═════╝  ╚═════╝
```

*by spaguetti-coder*

## Description

Ploydoo is an interactive CLI tool that sets up a complete Odoo development environment in minutes. It guides you through selecting your Odoo version, OCA modules, custom addons, PostgreSQL database, and Python environment -- then clones everything, configures it, and gets you ready to develop.

## Features

- Interactive TUI with keyboard navigation powered by Bubbletea
- Clone any Odoo version (16.0, 17.0, 18.0) with a single selection
- Browse and select from 30+ OCA community modules
- Optionally add your own custom addons repository with branch selection
- Spin up a PostgreSQL Docker container with your chosen version (14-17)
- Configure database credentials (user, password, database name)
- Detect and optionally install the correct Python version via pyenv
- Detect and optionally install Poetry via pipx
- Auto-generate `pyproject.toml` from Odoo's `requirements.txt`
- Set up a Poetry virtual environment with all dependencies
- Generate `odoo.conf` with all addons paths pre-configured
- Generate a `start.sh` script that initializes the database on first run

## Installation

```bash
brew tap alvarezperezda/tap && brew install ploydoo
```

## Usage

```bash
ploydoo
```

## Flow

1. **Installation path** -- Choose where Odoo and addons will be installed
2. **Odoo version** -- Select 16.0, 17.0, or 18.0
3. **OCA modules** -- Browse the list and toggle the community modules you need
4. **Custom addons** -- Optionally provide your own git repository URL and select a branch
5. **PostgreSQL version** -- Pick a PostgreSQL version for the Docker container
6. **Database configuration** -- Set the database user, password, and name
7. **Python environment** -- Review detected Python/Poetry status; optionally install missing tools
8. **Setup** -- Ploydoo clones repos, starts PostgreSQL, configures Python, and generates config files
9. **Done** -- Summary of results with `odoo.conf` and `start.sh` ready to use

## Requirements

- **git** -- for cloning repositories
- **docker** -- for running the PostgreSQL container
- **pyenv** (optional) -- for installing the required Python version automatically
- **poetry** (or auto-installed via pipx) -- for managing the Python virtual environment
- **Python 3.10+** -- required by Odoo 16.0/17.0 (3.12 for 18.0)

## License

MIT

## Made with

Go, [Bubbletea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss)
