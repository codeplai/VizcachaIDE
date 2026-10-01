## VizcachaIDE 1.0.0-rc1 — primera versión candidata

VizcachaIDE es un IDE de Go para principiantes, inspirado en Thonny, en español y en inglés.
Esta es la primera **versión candidata** de la 1.0: pruébala y cuéntanos qué falla antes de la
versión final. *(English notes: [notes-1.0.0-rc1.en.md](https://github.com/codeplai/VizcachaIDE/blob/v1.0.0-rc1/docs/release/notes-1.0.0-rc1.en.md).)*

![VizcachaIDE detenido en un punto de interrupción](https://raw.githubusercontent.com/codeplai/VizcachaIDE/v1.0.0-rc1/docs/images/debugger.es.png)

### Lo más destacado
- **Un depurador de verdad, basado en Delve.** Puntos de interrupción, paso sobre / adentro /
  afuera, ejecutar hasta el cursor, variables que se pueden desplegar, pila de llamadas y
  goroutines. El depurador simulado de la 0.1 desaparece.
- **Un Asistente que explica los errores de Go** con palabras sencillas, en español o en
  inglés: 25 tipos de errores de compilación, panics y avisos de `go vet` frecuentes, con una
  sugerencia para arreglarlos, el mensaje original y un enlace a la línea.
- **Inteligencia de código con gopls:** autocompletado (Ctrl+Space), errores subrayados mientras
  escribes, documentación al pasar el ratón, Ctrl+clic para ir a la definición, ayuda con los
  parámetros y panel Outline.
- **Un editor mejor:** temas, gofmt al guardar, buscar y reemplazar, ir a línea, comentar,
  zoom y archivos recientes.
- **Módulos de Go, argumentos y entrada por teclado** para tus programas, un panel Files y un
  diálogo *Go Modules* para `go mod init`, `go get` y `go mod tidy`.
- **Paquetes "full" con Go incluido**, para que quien empieza no tenga que instalar nada más.

La lista completa está en el
[CHANGELOG](https://github.com/codeplai/VizcachaIDE/blob/v1.0.0-rc1/CHANGELOG.md#historial-de-cambios-español).

### Descargas

| Archivo | Contenido |
|---|---|
| `VizcachaIDE-<versión>-windows-x64-full-setup.exe` | Instalador de Windows: IDE + Go 1.25.14, Delve 1.27.2 y gopls 0.21.1 |
| `VizcachaIDE-<versión>-windows-x64-lite-setup.exe` | Instalador de Windows: sólo el IDE (usa el Go de tu sistema) |
| `VizcachaIDE-<versión>-windows-x64-{full,lite}-portable.zip` | Windows, sin instalación |
| `VizcachaIDE-<versión>-macos-arm64-{full,lite}.dmg` | macOS con Apple Silicon (**vista previa**) |
| `VizcachaIDE-<versión>-macos-x86_64-{full,lite}.dmg` | macOS con Intel (**vista previa**) |
| `VizcachaIDE-<versión>-linux-x86_64-{full,lite}.AppImage` | Linux x86_64 (**vista previa**) |
| `SHA256SUMS.txt` | Sumas de verificación de todos los archivos |
| `NOTICE.md` | Aviso de licencia del programa distribuido |

**¿Cuál elijo?** Si estás empezando con Go, elige **full**. Si ya tienes Go instalado, **lite**
pesa mucho menos (unos 30 MB en lugar de unos 110 MB comprimido, en Windows). Con lite, instala
Delve y gopls para depurar y para la inteligencia de código:
`go install github.com/go-delve/delve/cmd/dlv@latest` y
`go install golang.org/x/tools/gopls@latest`.

### Requisitos del sistema
- **Windows:** Windows 10 u 11 de 64 bits (x64).
- **macOS:** macOS 11 Big Sur o posterior, con Apple Silicon o Intel. La app no está
  notarizada: la primera vez, haz clic derecho sobre ella y elige **Abrir**.
- **Linux:** x86_64 con glibc 2.35 o posterior (por ejemplo, Ubuntu 22.04 o superior) y las
  bibliotecas de Qt habituales (`libxkbcommon-x11-0`, `libxcb-*`, `libegl1`, `libfontconfig1`).
  Dale permiso de ejecución con `chmod +x` y ábrelo.
- **Sólo lite:** Go instalado (se recomienda 1.21 o posterior; se probó con la 1.25).
- Espacio en disco en Windows, una vez instalado: unos 70 MB (lite) o 310 MB (full).

### Licencia
El código fuente de VizcachaIDE es MIT. Estos binarios incluyen además PyQt5, que es GPLv3, así
que **el programa distribuido, en su conjunto, queda bajo la GPLv3**. La lista completa de
componentes incluidos y sus licencias está en `NOTICE.md`. La GPL se aplica al IDE, no a los
programas de Go que escribas con él.

### Limitaciones conocidas
- **Los paquetes de macOS y Linux son una vista previa**: los genera el workflow de release,
  pero todavía no se han probado en máquinas reales. Avísanos de cualquier cosa que no funcione.
- El instalador de Windows no se había probado antes de esta versión candidata (sí la
  aplicación y el zip portable).
- Sin firma de código: SmartScreen de Windows puede avisar de un editor desconocido.
- La traducción al español aún no está completa y el idioma sigue al del sistema operativo
  (todavía no hay selector en Opciones).
- Un programa en depuración no puede leer del teclado (stdin).
- La disposición inicial de los paneles todavía se está ajustando; el panel Asistente puede
  empezar pequeño (arrastra su borde o despégalo).

### Gracias
A los proyectos que hacen posible VizcachaIDE: Thonny (la inspiración), Delve, gopls, Go,
PyQt5 y Qt.
