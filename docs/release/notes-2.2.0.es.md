## VizcachaIDE 2.2.0 — llega Python

VizcachaIDE es un IDE para principiantes en inglés y español, inspirado en Thonny. La versión 2.2
añade **Python** como segundo lenguaje, con la misma experiencia que ya tenía Go. *(Release notes in
English: [notes-2.2.0.en.md](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.2.0/docs/release/notes-2.2.0.en.md).)*

![Depurando un programa de Python que lee el teclado](https://raw.githubusercontent.com/codeplai/VizcachaIDE/wails-v2.2.0/docs/wails/qa-2.2/es-py-05-debug-input.png)

### Novedades
- **Ejecuta Python con F5.** Tu programa corre en una terminal de verdad: `input()` lee lo que
  escribes en Salida y nunca se pierde lo que imprime antes de caerse.
- **Errores explicados en tu idioma.** Unos 30 errores comunes de Python (un nombre que no existe, la
  sangría equivocada, falta de dos puntos, sumar texto y números, dividir entre cero…) se explican con
  qué pasó, por qué y cómo arreglarlo. El mensaje original se queda como lo escribió Python.
- **Un depurador de verdad.** Puntos de interrupción, paso a paso, variables que se iluminan al
  cambiar, la pila de llamadas, parada en la línea de un error no capturado, y puedes **escribir
  respuestas mientras depuras**.
- **Ayuda mientras escribes.** Problemas subrayados al teclear, sugerencias, documentación al pasar
  el ratón e ir a la definición.
- **Formato al guardar**, una **consola de Python** que recuerda tus variables y un diálogo
  **Paquetes** para instalar bibliotecas con pip.
- **Elige tus lenguajes.** El asistente de primer arranque pregunta qué lenguajes de programación
  usarás; los menús y Ajustes muestran sólo esos.

### Descargas

| Archivo | Contenido |
|---|---|
| `VizcachaIDE-2.2.0-windows-amd64-full-setup.exe` | Instalador de Windows: IDE + Go y Python, con sus depuradores y ayudantes de código |
| `VizcachaIDE-2.2.0-windows-amd64-full-python-setup.exe` | Instalador de Windows: IDE + Python |
| `VizcachaIDE-2.2.0-windows-amd64-full-go-setup.exe` | Instalador de Windows: IDE + Go (el antiguo "full") |
| `VizcachaIDE-2.2.0-windows-amd64-lite-setup.exe` | Instalador de Windows: sólo el IDE (usa el Go o el Python de tu sistema) |
| `VizcachaIDE-2.2.0-windows-amd64-<variante>-portable.zip` | Windows, sin instalar |
| `VizcachaIDE-2.2.0-darwin-*`, `VizcachaIDE-2.2.0-linux-*` | macOS y Linux (**vista previa**) |
| `SHA256SUMS-*.txt` | Sumas de verificación de cada archivo |

**¿Cuál elijo?** Para un curso de Python, **full-python** (unos 70 MB comprimido en Windows); para Go
y Python, **full**. Con **lite**, instala Python 3.10 o más nuevo y ejecuta
`python -m pip install debugpy "python-lsp-server[pyflakes]" ruff`. Esta versión se probó en
Windows 10; las de macOS y Linux salen del CI y no se han probado a mano.

Consulta el [CHANGELOG](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.2.0/CHANGELOG.md)
para la lista completa.
