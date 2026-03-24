package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateOdooConf creates an odoo.conf file with the given addons paths.
func GenerateOdooConf(baseDir string, ocaModules []string, includeCustomAddons bool, odooVersion, dbUser, dbPassword, dbName, dbPort string) error {
	paths := []string{
		"./odoo/addons",
		"./odoo/odoo/addons",
	}

	for _, mod := range ocaModules {
		paths = append(paths, fmt.Sprintf("./addons/%s", mod))
	}

	if includeCustomAddons {
		paths = append(paths, "./addons/custom_addons")
	}

	conf := fmt.Sprintf(`[options]
addons_path = %s
db_host = localhost
db_port = %s
db_user = %s
db_password = %s
db_name = %s
admin_passwd = admin
http_port = 8069
log_level = info
`, strings.Join(paths, ","), dbPort, dbUser, dbPassword, dbName)

	confPath := filepath.Join(baseDir, "odoo.conf")
	if err := os.WriteFile(confPath, []byte(conf), 0644); err != nil {
		return err
	}

	return GenerateStartScript(baseDir, dbUser, dbPassword, dbName, dbPort)
}

// GenerateStartScript creates a start.sh script that initializes the DB on first run and starts Odoo.
func GenerateStartScript(baseDir, dbUser, dbPassword, dbName, dbPort string) error {
	script := fmt.Sprintf(`#!/bin/bash

BASEDIR="$(cd "$(dirname "$0")" && pwd)"
CONF="$BASEDIR/odoo.conf"
ODOO_BIN="$BASEDIR/odoo/odoo-bin"
DB_NAME="%s"
DB_HOST="localhost"
DB_PORT="%s"
DB_USER="%s"
DB_PASSWORD="%s"
INIT_FLAG="$BASEDIR/.db_initialized"

# Work from BASEDIR so poetry finds its environment
cd "$BASEDIR"

if [ ! -f "$ODOO_BIN" ]; then
    echo "Error: odoo-bin not found at $ODOO_BIN"
    exit 1
fi

# First run: initialize database without demo data
if [ ! -f "$INIT_FLAG" ]; then
    echo "Initializing database '$DB_NAME' (without demo data)..."

    # Wait for PostgreSQL to be ready
    echo "Waiting for PostgreSQL..."
    for i in $(seq 1 30); do
        if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT 1" > /dev/null 2>&1; then
            echo "PostgreSQL is ready."
            break
        fi
        if [ "$i" -eq 30 ]; then
            echo "Error: PostgreSQL is not available after 30 seconds."
            exit 1
        fi
        sleep 1
    done

    # Create database if it doesn't exist
    if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -tc \
        "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1; then
        echo "Creating database '$DB_NAME'..."
        PGPASSWORD="$DB_PASSWORD" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"
    fi

    echo "Initializing Odoo (this may take a few minutes)..."
    poetry run python "$ODOO_BIN" -c "$CONF" -d "$DB_NAME" -i base --without-demo=all --stop-after-init --no-http
    if [ $? -ne 0 ]; then
        echo "Error: Odoo initialization failed."
        exit 1
    fi

    touch "$INIT_FLAG"
    echo "Database initialized successfully."
fi

echo "Starting Odoo..."
exec poetry run python "$ODOO_BIN" -c "$CONF"
`, dbName, dbPort, dbUser, dbPassword)

	scriptPath := filepath.Join(baseDir, "start.sh")
	return os.WriteFile(scriptPath, []byte(script), 0755)
}
