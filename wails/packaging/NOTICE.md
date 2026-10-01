VizcachaIDE (Wails variant) - LICENSE NOTICE / AVISO DE LICENCIAS
==================================================================

English
-------
The VizcachaIDE source code is released under the MIT License
(Copyright (c) 2025-2026 Marks Calderon - Codeplai Games). The full text follows this notice.

Unlike the 1.x series (PyQt5, GPLv3), the Wails variant contains NO GPL code, so the
distributed program is licensed under the MIT License. It bundles or links the following
third-party software, all under permissive licenses:

  Component                        License        Where
  -------------------------------  -------------  -----------------------------------
  Wails v2 (Lea Anthony et al.)    MIT            linked into the executable
  Svelte, CodeMirror 6, Vite       MIT            compiled into the frontend assets
  Go toolchain *  (The Go Authors) BSD-3-Clause   toolchain/go/LICENSE
  Delve (dlv) *   (Derek Parker)  MIT            toolchain/licenses/dlv-LICENSE.txt
  gopls *         (The Go Authors) BSD-3-Clause   toolchain/licenses/gopls-LICENSE.txt
  Microsoft WebView2 bootstrapper  Microsoft      Windows installer only; it installs the
                                                  WebView2 runtime if the PC lacks it
  (* only in the "full" variant; the "lite" variant uses the Go installed on your system)

The platform web engine (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux) is
provided by the operating system and is not redistributed in the macOS and Linux packages.

This program is provided WITHOUT ANY WARRANTY, to the extent permitted by law.

Espanol
-------
El codigo fuente de VizcachaIDE se publica bajo la licencia MIT
(Copyright (c) 2025-2026 Marks Calderon - Codeplai Games). El texto completo sigue a este aviso.

A diferencia de la serie 1.x (PyQt5, GPLv3), la variante Wails NO contiene codigo GPL, asi que
el programa distribuido tiene licencia MIT. Incluye o enlaza el siguiente software de terceros,
todo con licencias permisivas:

  Wails v2 (MIT), Svelte / CodeMirror 6 / Vite (MIT), y solo en la variante "full":
  Go (BSD-3-Clause), Delve (MIT) y gopls (BSD-3-Clause), con sus licencias en toolchain/.
  El instalador de Windows incluye el bootstrapper de Microsoft WebView2, que instala el
  runtime si el equipo no lo tiene. La variante "lite" usa el Go instalado en tu sistema.

El motor web (WebView2 en Windows, WKWebView en macOS, WebKitGTK en Linux) lo aporta el
sistema operativo y no se redistribuye en los paquetes de macOS y Linux.

Este programa se ofrece SIN NINGUNA GARANTIA, en la medida permitida por la ley.
