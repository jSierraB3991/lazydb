# LazyDB 🐘

**LazyDB** es un cliente de base de datos ligero, rápido e intuitivo para la terminal (TUI - Terminal User Interface) escrito en Go. Utiliza las librerías `tview` y `tcell` para ofrecer una interfaz fluida e interactiva directamente desde tu terminal, permitiéndote explorar y gestionar bases de datos PostgreSQL de forma ágil sin dependencias externas pesadas.

---

## 🎨 Vista Previa de la Interfaz

```text
+-----------------------------------------------------------------------------+
|  Conexiones                     |  Datos: public.usuarios                   |
| 🐘 postgres_local               |-------------------------------------------|
| 🐘 production_replica           | id   | nombre   | email       | rol       |
|                                 |------+----------+-------------+-----------|
|                                 | 1    | Lelouch  | le@b.com    | admin     |
|                                 | 2    | CC       | cc@b.com    | user      |
|---------------------------------| 3    | Kallen   | ka@b.com    | user      |
|  Schema / Tablas                |      |          |             |           |
| 📦 postgres_local               |      |          |             |           |
|  📁 public                      |      |          |             |           |
|   🗃️ usuarios                  |      |          |             |           |
|   🗃️ ordenes                   |      |          |             |           |
|                                 |      |          |             |           |
+-----------------------------------------------------------------------------+
| λ public.usuarios - 3 filas  |  Ctrl+C: Salir                               |
+-----------------------------------------------------------------------------+
```

---

## ✨ Características Principales

- 🖥️ **TUI Moderna y Colorida**: Interfaz interactiva nativa de terminal con colores y bordes dinámicos que indican el foco actual.
- 🗃️ **Gestor de Conexiones Integrado**: Añade, lista y elimina perfiles de conexión de base de datos desde la propia aplicación.
- 💾 **Persistencia de Configuración**: Guarda de forma segura las credenciales de tus bases de datos localmente en formato JSON.
- 📦 **Navegador de Esquemas y Tablas**: Visualización jerárquica tipo árbol con esquemas y tablas asociadas.
- 📊 **Visualizador de Datos**: Visualiza registros en formato cuadrícula (limitado automáticamente a 500 filas por consulta para optimizar rendimiento).
- 🗑️ **Acciones Rápidas**: Elimina registros seleccionados directamente de la base de datos con un solo botón (con modal de confirmación de seguridad).
- 🔔 **Barra de Estado**: Mensajes informativos en tiempo real sobre errores, conexiones y comandos.

---

## 🛠️ Requisitos Previos

Para compilar y ejecutar LazyDB, necesitas:

- **Go** (versión 1.26 o superior) instalado en tu sistema.
- Acceso a una base de datos **PostgreSQL** para probar las conexiones.

---

## 🚀 Instalación y Compilación

1. **Clona el repositorio:**
   ```bash
   git clone https://github.com/jsierrab3991/lazydb.git
   cd lazydb
   ```

2. **Compila la aplicación:**
   Puedes compilar el binario ejecutable directamente con:
   ```bash
   go build -o lazydb main.go
   ```

3. **Ejecuta la aplicación:**
   ```bash
   ./lazydb
   ```

---

## ⚙️ Configuración y Almacenamiento

### Directorio de Configuración
LazyDB guarda las credenciales y perfiles de tus conexiones en tu directorio personal de configuración en formato JSON:
- **Ruta:** `~/.config/lazydb/connections.json`

El archivo de configuración tiene la siguiente estructura:
```json
[
  {
    "name": "Local DB",
    "type": "postgres",
    "host": "localhost",
    "port": "5432",
    "user": "mi_usuario",
    "password": "mi_password",
    "database": "mi_db"
  }
]
```

### Variables de Entorno
- **`LAZYDB_PORT`**: Puedes configurar esta variable opcional. Si se define, LazyDB validará que sea un número válido al iniciar la aplicación.

---

## ⌨️ Atajos de Teclado y Navegación

| Tecla / Combinación | Acción | Contexto |
|---|---|---|
| `Tab` | Cambiar el foco al siguiente panel en sentido horario | Global |
| `Shift + Tab` (Backtab) | Cambiar el foco al panel anterior en sentido antihorario | Global |
| `Espacio` | Abrir el formulario para agregar una **Nueva Conexión** | Pantalla Principal / Menú |
| `Intro` (Enter) | Conectar a la BD seleccionada / Cargar datos de la tabla / Expandir nodo del árbol | Listas y Árbol de Esquemas |
| `Supr` (Delete) o `Retroceso` (Backspace) | Eliminar la conexión seleccionada | Lista de Conexiones |
| `Supr` (Delete) | Eliminar la fila actualmente seleccionada | Vista de Datos (Tabla) |
| `Ctrl + C` | Salir de la aplicación inmediatamente | Global |

---

## 📂 Estructura del Proyecto

- `main.go`: Punto de entrada de la aplicación, inicializa la interfaz de usuario y valida el entorno.
- `app/`: Directorio principal del código fuente:
  - `app.go`: Maneja el ciclo de vida de la TUI, inicialización de componentes (`tview.List`, `tview.TreeView`, `tview.Table`), y flujo de diálogos modales.
  - `connection.go`: Define la estructura `Connection` y métodos para generar el DSN (Data Source Name) de PostgreSQL y cargar conexiones locales.
  - `db.go`: Contiene la lógica para guardar perfiles y ejecutar sentencias SQL (como `DELETE` para registros y consultas de metadatos de esquema).
  - `libs.go`: Utilidades, constantes de la interfaz, colores y rutas de configuración.

---

## 🤝 Contribuciones

Las contribuciones son bienvenidas. Siéntete libre de abrir un *Issue* o enviar un *Pull Request* para mejorar LazyDB.

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT.
