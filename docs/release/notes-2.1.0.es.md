## VizcachaIDE 2.1.0 — lista para más lenguajes

VizcachaIDE es un IDE para principiantes en inglés y español, inspirado en Thonny. La versión 2.1
reorganiza el interior del IDE en torno a **perfiles de lenguaje**: Go funciona exactamente igual
que antes, y Python y C++ ya se pueden añadir sin cambiar el resto del IDE. *(Release notes in
English: [notes-2.1.0.en.md](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.1.0/docs/release/notes-2.1.0.en.md).)*

![Configuración, pestaña Herramientas, con las herramientas de Go encontradas](https://raw.githubusercontent.com/codeplai/VizcachaIDE/wails-v2.1.0/docs/wails/qa-2.1/persist-02-tools.png)

### Novedades
- **Nuevo archivo de…** en el menú Archivo: Go, Python o C++, cada uno con su plantilla inicial.
  En Ajustes eliges el lenguaje de los archivos nuevos.
- **Herramientas por lenguaje.** Ajustes → Herramientas muestra las herramientas de cada lenguaje,
  dónde se encontraron y su versión.
- **Avisos más claros cuando falta una herramienta.** El aviso dice cuál falta y ofrece *Instalar*
  o *Copiar el comando*, además de *Elegir en Ajustes*.
- **Más ligero en segundo plano.** El ayudante de código (gopls) se detiene tras 5 minutos sin
  archivos abiertos y vuelve a arrancar solo.

### Para Go, igual que antes
Ejecutar con entrada por teclado y argumentos, Detener, los errores explicados en tu idioma, las
sugerencias y los problemas en vivo, el formato con gofmt al guardar, los módulos de Go y el
depurador real funcionan como en la 2.0. Lo comprobamos con 34 pasos automáticos en inglés y
español.

### Conviene saber
- **Tus ajustes se conservan.** Las rutas de herramientas de un `settings.json` de la 2.0 se
  convierten la primera vez que arranca la 2.1.
- **Python y C++ todavía no se ejecutan.** Puedes crear y editar sus archivos, pero al ejecutarlos
  el IDE avisa de que aún no está disponible. Su soporte es el siguiente paso.
- El diálogo "Módulos de Go" ahora se llama **Paquetes**. Para Go tiene las mismas acciones.

### Descargas

| Archivo | Contenido |
|---|---|
| `VizcachaIDE-2.1.0-windows-amd64-full-setup.exe` | Instalador de Windows: IDE + Go, Delve y gopls |
| `VizcachaIDE-2.1.0-windows-amd64-lite-setup.exe` | Instalador de Windows: solo el IDE (usa el Go de tu sistema) |
| `VizcachaIDE-2.1.0-windows-amd64-{full,lite}-portable.zip` | Windows, sin instalar |
| `VizcachaIDE-2.1.0-darwin-{arm64,amd64}-{full,lite}.dmg` | macOS (**vista previa**, sin notarizar) |
| `VizcachaIDE-2.1.0-linux-amd64-{full,lite}.{AppImage,tar.gz}` | Linux (**vista previa**) |
| `SHA256SUMS-*.txt` | Sumas de verificación de cada archivo |

**¿Cuál elijo?** Si empiezas con Go, toma **full**. Si ya tienes Go, **lite** es mucho más
pequeño. Esta versión se probó en Windows 10; las de macOS y Linux salen del CI y no se han
probado a mano.

Consulta el [CHANGELOG](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.1.0/CHANGELOG.md)
para la lista completa.
