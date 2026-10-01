---
title: VizcachaIDE
description: El IDE de Go para quien recién empieza. Escribe, ejecuta, entiende tus errores y depura paso a paso, en español o en inglés.
image: https://raw.githubusercontent.com/codeplai/VizcachaIDE/main/vizcachaidelogo.png
---

![VizcachaIDE by codeplai](https://raw.githubusercontent.com/codeplai/VizcachaIDE/main/vizcachaidelogo.png)

# VizcachaIDE

**El IDE de Go para quien recién empieza.** Escribe tu programa, ejecútalo con un botón, entiende por qué falla y míralo funcionar paso a paso. Todo en una sola ventana, en español o en inglés.

VizcachaIDE se inspira en [Thonny](https://thonny.org), el IDE con el que miles de personas aprenden Python, y lleva esa misma idea a Go: menos botones, más claridad y explicaciones pensadas para aprender.

## Para quién es

- **Estudiantes** que dan sus primeros pasos en programación con Go.
- **Docentes** que necesitan una herramienta sencilla para el aula, que funcione en español y sin configuraciones complicadas.
- **Cualquier persona** que quiera aprender Go sin pelearse primero con terminales, variables de entorno y extensiones.

## Qué puedes hacer

### Escribir y ejecutar
- Pulsa **Ejecutar (F5)** y mira la salida de tu programa al instante.
- Escribe en la consola cuando tu programa te pida datos por teclado.
- Pasa argumentos a tu programa y trabaja con proyectos que usan `go.mod`.
- Tu código se ordena solo al guardar, con el formato estándar de Go (gofmt).

### Entender tus errores
- Cuando Go encuentra un problema, el **Asistente** te explica **qué pasó y cómo arreglarlo**, en tu idioma.
- Reconoce los 25 errores más comunes de quien empieza: variables sin usar, imports de más, tipos que no encajan, índices fuera de rango, mapas sin inicializar, bloqueos entre goroutines y más.
- El mensaje original de Go siempre está a la vista, con un botón para buscarlo en internet. Así aprendes a leer los errores reales.
- Los errores se subrayan **mientras escribes**, antes de ejecutar.

### Ver tu programa paso a paso
- Haz clic junto a un número de línea para poner un **punto de interrupción** y pulsa **Depurar (F6)**.
- Avanza con botones que hablan claro: **Siguiente línea**, **Entrar en la función**, **Salir de la función**.
- Mira el valor de tus variables en cada paso. La que acaba de cambiar se resalta, para que veas qué hizo la última línea.
- Descubre **cómo llegaste ahí** (la pila de llamadas) y qué hace cada goroutine.

### Escribir más rápido
- Autocompletado inteligente de Go, con la documentación de cada función.
- Ayuda con los parámetros mientras escribes una llamada.
- Ctrl+clic para ir a donde se define una función.
- Buscar y reemplazar, ir a una línea, zoom, tema claro u oscuro.

## Pensado para aprender

- **Español e inglés** en toda la interfaz y en las explicaciones de errores. Detecta el idioma de tu sistema y puedes cambiarlo cuando quieras.
- **Un botón principal.** Ejecutar es lo más visible de la ventana; lo demás aparece cuando lo necesitas.
- **Letra muy legible.** Usa Atkinson Hyperlegible, una tipografía diseñada para que caracteres como 0 y O, o 1, l e I, no se confundan.
- **Todo incluido.** La versión completa trae Go, el depurador Delve y gopls: instalas VizcachaIDE y ya puedes programar.

## Descarga

VizcachaIDE es **gratis y de código abierto**.

| Versión | Qué incluye | Para quién |
|---|---|---|
| **Completa** | VizcachaIDE + Go + Delve + gopls | Si no tienes Go instalado (recomendada para empezar) |
| **Ligera** | Solo VizcachaIDE | Si ya tienes Go instalado |

**Descargas y código fuente:** [github.com/codeplai/VizcachaIDE](https://github.com/codeplai/VizcachaIDE)

**Sistemas:**
- **Windows 10 y 11:** instalador sin permisos de administrador y versión portable.
- **macOS y Linux:** versión preliminar.

> **Primera vez en Windows:** como el instalador todavía no tiene firma digital, Windows puede mostrar "Windows protegió tu PC". Haz clic en **Más información → Ejecutar de todas formas**.

## Estado del proyecto

VizcachaIDE está en **versión candidata (release candidate)**: ya se puede usar y estamos puliendo detalles antes de la versión final.

- **Nueva edición 2.0:** interfaz rediseñada, más ligera y rápida, y con un depurador que explica cada paso.
- **Edición clásica 1.x:** sigue disponible mientras la 2.0 llega a su versión final.

## Hecho en Perú

VizcachaIDE es un proyecto de **[Codeplai Games](https://codeplai.pe)**, creado por Marks Calderon. Su nombre viene de la **vizcacha**, el roedor de los Andes que vive entre las rocas de la sierra: pequeño, curioso y siempre atento.

¿Tienes ideas, encontraste un error o quieres usarlo en tu clase? Escríbenos a **hola@codeplai.pe** o abre un *issue* en GitHub.

---

<small>VizcachaIDE se distribuye bajo licencia MIT e incluye Go, Delve y gopls, cada uno con su propia licencia de código abierto.</small>
