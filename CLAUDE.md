## Ploydoo CLI

CLI interactivo en Go con Bubbletea para montar entornos de desarrollo Odoo completos.

### Stack
- **Go** + **Bubbletea** (TUI) + **Bubbles** (componentes) + **Lipgloss** (estilos)
- Git clone via `os/exec`, Docker para PostgreSQL, Poetry + pyenv para Python

### Flujo del CLI

1. **Ruta de instalación** — input de texto, soporta `~/`, crea el directorio si no existe
2. **Versión de Odoo** — selector (16.0, 17.0, 18.0) → `git clone --depth 1` de `odoo/odoo`
3. **Módulos OCA** — multi-select con scroll (42 módulos), `space` toggle, `a` all, `n` none
4. **Rama alventia_modules** — lista ramas remotas de `git@github.com:daperez89/alventia_modules.git` y permite seleccionar cuál clonar
5. **Versión PostgreSQL** — selector (14, 15, 16, 17)
6. **Configuración BD** — campos editables: usuario, contraseña, nombre BD (defaults: odoo/odoo/odoo-{version})
7. **Entorno Python** — detecta Python compatible (3.10 para Odoo 16-17, 3.12 para Odoo 18), ofrece instalar con pyenv si falta
8. **Ejecución** — progreso con spinner por cada tarea:
   - Clone Odoo
   - Clone módulos OCA seleccionados (en `addons/`)
   - Clone alventia_modules (rama seleccionada, en `addons/`)
   - Contenedor Docker PostgreSQL (`odoo-postgres-{version}`, `-d --rm`, puerto 5432)
   - Instalación Python via pyenv (si se confirmó)
   - Generación `pyproject.toml` (parsea `requirements.txt`, filtra deps Windows, evalúa markers de python_version, deduplica)
   - `poetry env use` + `poetry install` (`package-mode = false`)
9. **Generación de archivos**:
   - `odoo.conf` — addons_path con todas las rutas, config BD, sin logfile (errores visibles en consola)
   - `start.sh` — script de arranque que:
     - Espera a PostgreSQL (hasta 30s)
     - Crea la BD si no existe
     - Primera ejecución: `poetry run python odoo-bin -i base --without-demo=all --stop-after-init --no-http`
     - Marca `.db_initialized` como flag
     - Siguientes ejecuciones: `poetry run python odoo-bin -c odoo.conf`

### Estructura del proyecto

```
main.go                        — punto de entrada
internal/
  tui/
    model.go                   — modelo Bubbletea, máquina de estados, lógica de cada paso
    styles.go                  — estilos lipgloss
    logo.go                    — logo ASCII "PLOYDOO" + firma "by spaguetti-coder"
    path.go                    — vista input ruta de instalación
    version.go                 — vista selector versión Odoo
    modules.go                 — vista multi-select módulos OCA con scroll
    alventia.go                — vista selector rama alventia_modules
    postgres.go                — vista selector versión PostgreSQL
    dbconfig.go                — vista config BD (usuario, contraseña, nombre)
    python.go                  — vista confirmación entorno Python
    progress.go                — vista progreso clonación con spinner
  git/
    clone.go                   — git clone (Odoo, OCA, alventia) + listado ramas remotas
  docker/
    postgres.go                — docker run PostgreSQL
  python/
    setup.go                   — pyenv install, parseo requirements.txt, pyproject.toml, poetry setup
  config/
    odoo.go                    — generación odoo.conf + start.sh
```

### Notas técnicas
- Todos los clones usan `--depth 1` para ahorrar espacio
- El parseo de `requirements.txt` filtra paquetes Windows-only (`pywin32`, `pypiwin32`), evalúa markers `python_version` y `sys_platform`, y deduplica por nombre
- La lista de módulos OCA y ramas alventia son scrollables, adaptándose al tamaño de la terminal via `tea.WindowSizeMsg`

### Instrucciones para el agente
- Usa skills de engram para almacenar memoria y consumir menos tokens