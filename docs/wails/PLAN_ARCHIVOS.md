# Plan: manejo de archivos al estilo Thonny (VizcachaIDE 2.0)

**Decisiones del dueño (2026-10-01):** «Eliminar» manda a la **Papelera de reciclaje**; acceso con **menú Archivo + 3 botones** (Nuevo, Abrir, Guardar).

## Qué falta hoy

| Acción | Thonny | VizcachaIDE 2.0 hoy |
|---|---|---|
| Archivo nuevo (Ctrl+N) | Menú Archivo y botón en la barra | Solo desde el asistente de primer arranque |
| Abrir archivo (Ctrl+O) | Menú y botón | **No existe**: solo «Abrir carpeta…» en «Más» |
| Abrir carpeta | Panel Archivos | En «Más» y en el panel vacío |
| Archivos recientes | Menú Archivo | En «Más» (recién agregado) |
| Guardar (Ctrl+S) | Menú y botón | Solo con el atajo; **un archivo nuevo no se puede guardar** |
| Guardar como (Ctrl+Shift+S) | Sí | **No existe** |
| Guardar todo | Sí | No existe |
| Cerrar (Ctrl+W) y cerrar todo | Sí | Solo con la × de cada pestaña |
| Crear, renombrar o eliminar desde el panel Archivos | Menú contextual | **No existe** |
| Mostrar en el Explorador | Menú contextual | No existe |

El hueco más grave: un programa empezado con «archivo en blanco» se puede ejecutar, pero **no se puede guardar**.

## Diseño de la interfaz

1. **Menú «Archivo» visible** en la barra de título, a la izquierda, junto al logo. Es lo que espera un principiante, y Thonny lo tiene en el mismo lugar. Contenido y atajos:
   - Nuevo archivo (Ctrl+N)
   - Abrir archivo… (Ctrl+O)
   - Abrir carpeta…
   - Abrir reciente ▸ (se mueve aquí desde «Más»)
   - Guardar (Ctrl+S)
   - Guardar como… (Ctrl+Shift+S)
   - Guardar todo
   - Cerrar (Ctrl+W)
   - Cerrar todo

   «Más» queda solo con: Módulos de Go, Configuración y Acerca de.
2. **Tres botones de icono** a la izquierda de Ejecutar: Nuevo, Abrir y Guardar, con ayuda emergente y su atajo. «Guardar» se ve desactivado cuando no hay cambios. Ejecutar sigue siendo el botón principal.
3. **Panel Archivos**:
   - En la cabecera, botones para nuevo archivo, nueva carpeta y actualizar.
   - Clic derecho sobre un archivo o carpeta: Nuevo archivo aquí, Nueva carpeta, Renombrar (F2), Eliminar (Supr), Mostrar en el Explorador, Copiar ruta.
   - Al crear un archivo o carpeta, el nombre se escribe en la misma fila del árbol.
4. **Archivo nuevo**:
   - Ctrl+N abre una pestaña «sin título» con la plantilla mínima de Go (`package main` + `func main`), como hoy.
   - Si hay una carpeta abierta, el nombre se pide en el árbol.
5. **Guardar un archivo sin título**: Ctrl+S abre «Guardar como». Al guardarlo, la pestaña pasa a ser un archivo real; desde ahí funcionan gopls, el vigilante de cambios y los recientes. Cerrar una pestaña sin título con cambios ofrece «Guardar», que también abre «Guardar como».
6. **Eliminar**:
   - Pide confirmación con la acción en el botón: «¿Eliminar *main.go*?» → **Eliminar** / **Cancelar**.
   - Si el archivo está abierto, se cierra su pestaña.

## Textos (EN / ES)

Siguen [UX_COPY.md](UX_COPY.md): tuteo y verbos en los botones.

- *File* / *Archivo*
- *New file* / *Nuevo archivo*
- *Open file…* / *Abrir archivo…*
- *Save as…* / *Guardar como…*
- *Save all* / *Guardar todo*
- *Close* / *Cerrar*
- *Close all* / *Cerrar todo*
- *Rename* / *Renombrar*
- *Delete* / *Eliminar*
- *Show in Explorer* / *Mostrar en el Explorador* (en macOS, *Finder*)
- *Copy path* / *Copiar ruta*
- *New folder* / *Nueva carpeta*

Errores con la estructura qué pasó + cómo arreglarlo. Por ejemplo, «Ya existe un archivo llamado *main.go* en esta carpeta. Elige otro nombre.»

## Backend (Go)

`FilesService` gana estos métodos. Siguen las mismas reglas de arquitectura: diálogos nativos con el runtime de Wails, solo en `bridge`.

- `OpenFileDialog() (string, error)`: diálogo nativo con filtro «Archivos Go (*.go)» y «Todos».
- `SaveFileDialog(suggestedName, folder string) (string, error)`: diálogo nativo; agrega `.go` si falta.
- `CreateFile(path, text string) error` y `CreateFolder(path string) error`: fallan si ya existe.
- `Rename(from, to string) error`: falla si el destino existe.
- `Delete(path string) error`: manda a la Papelera. En Windows usa `SHFileOperationW` con `FOF_ALLOWUNDO` (shell32, vía golang.org/x/sys/windows). En macOS usa el Finder a través de `osascript`. En Linux usa `gio trash`. Si la Papelera no está disponible, devuelve un error claro y no borra.
- `RevealInExplorer(path string) error`: abre el Explorador con el archivo seleccionado (`explorer /select,` en Windows, `open -R` en macOS, `xdg-open` de la carpeta en Linux).

Las operaciones viven en `app` (casos de uso con validación de nombres) y el sistema de archivos en un adaptador. Llevan pruebas en carpetas temporales.

## Frontend

- `stores/fileCommands.ts`: nuevo, abrir, guardar como, guardar todo, cerrar todo.
- `stores/fileTreeCommands.ts`: crear, renombrar y eliminar en el árbol. Al renombrar o eliminar actualiza las pestañas abiertas, los buffers, gopls, el vigilante y los recientes.
- Componentes:
  - `shell/FileMenu.svelte`: bits-ui DropdownMenu.
  - `shell/FileButtons.svelte`.
  - `panels/FileTreeMenu.svelte`: bits-ui ContextMenu.
  - Edición del nombre en línea dentro de `FileTreeNode.svelte`.
- Atajos en `shell/shortcuts.ts`: Ctrl+N, Ctrl+O, Ctrl+Shift+S, Ctrl+W y F2/Supr en el árbol. Todos con `preventDefault`, para que WebView2 no abra ventanas ni cierre nada.
- Mock del bridge para `npm run dev` y pruebas con vitest.

## Ejecución

Dos agentes en paralelo, en copias separadas del repo, después de que terminen los agentes que siguen trabajando (consola y mejoras rápidas):

| Agente | Alcance |
|---|---|
| A: menú Archivo y guardar | Menú «Archivo», botones, Nuevo, Abrir archivo, Guardar como (incluye sin título), Guardar todo, Cerrar y Cerrar todo, atajos, métodos de diálogo del backend |
| B: panel Archivos | Botones de la cabecera, menú contextual, crear, renombrar, eliminar, mostrar en Explorador, copiar ruta, edición en línea, métodos de archivos del backend |

Para que no choquen, primero dejo listo en `main` el contrato: las firmas nuevas de `FilesService` en Go, `types.ts` y el mock.

## Verificación

- Pruebas de Go y vitest, lint y la prueba de arquitectura.
- Prueba en la app real (`wails dev`):
  - crear un archivo nuevo, ejecutarlo y guardarlo como `hola.go`;
  - abrir un archivo suelto;
  - renombrar y eliminar desde el árbol;
  - Ctrl+W y Cerrar todo con cambios sin guardar;
  - «Mostrar en el Explorador».
- Capturas en español e inglés, y a 1024 px de ancho.
- Actualizar el manual (artefacto) y `docs/wails/QA_WAILS.md`.
