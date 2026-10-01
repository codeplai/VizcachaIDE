# UX Copy: VizcachaIDE Wails (EN / ES)

Fuente de verdad de los textos de interfaz de la variante Wails. Los coders usan estos textos **tal cual**. Todo texto nuevo sigue las mismas reglas:
- **Claro:** sin jerga, o con el término técnico solo en la ayuda emergente.
- **Breve.**
- **Consistente:** un concepto, una palabra.
- **Útil:** cada error dice qué pasó, por qué y cómo arreglarlo.
- **Humano.**

**Tono:** un profesor paciente. Tranquilo cuando algo falla, breve cuando todo va bien y nunca condescendiente.

---

## Recommended Copy

### Acciones principales
| Elemento | English | Español | Ayuda emergente (EN / ES) |
|---|---|---|---|
| Ejecutar | **Run** | **Ejecutar** | Run your program (F5) / Ejecuta tu programa (F5) |
| Depurar | **Debug** | **Depurar** | Run step by step and see your variables (F6) / Ejecuta paso a paso y mira tus variables (F6) |
| Terminar depuración | **Stop debugging** | **Terminar depuración** | Shift+F5 |
| Detener programa | **Stop** | **Detener** | Stop the running program (Shift+F5) / Detén el programa en ejecución (Shift+F5) |
| Paso sobre | **Next line** | **Siguiente línea** | Runs this line without going inside functions (Step over) / Ejecuta esta línea sin entrar en funciones (Step over) |
| Paso adentro | **Go into function** | **Entrar en la función** | (Step into) |
| Paso afuera | **Leave function** | **Salir de la función** | (Step out) |
| Continuar | **Continue** | **Continuar** | Run until the next breakpoint (Shift+F6) / Ejecuta hasta el próximo punto de interrupción (Shift+F6) |
| Ejecutar hasta el cursor | **Run to here** | **Ejecutar hasta aquí** | Run until the line with the cursor (Ctrl+F10) / Ejecuta hasta la línea del cursor (Ctrl+F10) |

### Paneles
| Elemento | English | Español |
|---|---|---|
| Panel derecho | Assistant | Asistente |
| Salida | Output | Salida |
| Problemas | Problems | Problemas |
| Pila de llamadas | How you got here | Cómo llegaste aquí |
| Variables | Variables in {function}() | Variables en {function}() |
| Variable recién cambiada | just changed | acaba de cambiar |
| Estructura | Outline | Estructura |
| Archivos | Files | Archivos |
| Configuración | Settings | Configuración |

### Estados vacíos (qué es + por qué está vacío + cómo empezar)
| Lugar | English | Español | CTA |
|---|---|---|---|
| Archivos | No folder open. Open a folder to see your project's files here. | No hay ninguna carpeta abierta. Abre una para ver aquí los archivos de tu proyecto. | Open folder / Abrir carpeta |
| Salida | Nothing here yet. Press Run (F5) to see what your program prints. | Todavía no hay nada. Pulsa Ejecutar (F5) para ver lo que imprime tu programa. | — |
| Problemas | No problems found. Go checks your code as you type. | No hay problemas. Go revisa tu código mientras escribes. | — |
| Variables | Variables appear when the program pauses. Add a breakpoint and press Debug. | Las variables aparecen cuando el programa se pausa. Pon un punto de interrupción y pulsa Depurar. | — |
| Estructura | Functions and types of this file appear here once Go has read it. | Las funciones y tipos de este archivo aparecen aquí cuando Go termina de leerlo. | — |
| Assistant (consejos) | **Press Run or F5** to run your program. · **Something wrong?** You'll see what happened and how to fix it here, in your language. · **To watch your program step by step**, click next to a line number and press Debug. | **Pulsa Ejecutar o F5** para correr tu programa. · **¿Algo falla?** Aquí verás qué pasó y cómo arreglarlo, en tu idioma. · **Para ver tu programa paso a paso**, haz clic junto a un número de línea y pulsa Depurar. | — |

### Mensajes de ejecución
| Situación | English | Español |
|---|---|---|
| Inicio | ▶ Running {file}… | ▶ Ejecutando {file}… |
| Éxito | ✓ Finished in {seconds} s | ✓ Terminó bien en {seconds} s |
| Código de salida ≠ 0 | ✗ Your program ended with code {code}. | ✗ Tu programa terminó con código {code}. |
| Fallo al compilar | ✗ Your program didn't run: Go found {count, plural, one {# problem} other {# problems}}. See the explanation on the right. | ✗ No se pudo ejecutar: Go encontró {count, plural, one {# problema} other {# problemas}}. Mira la explicación a la derecha. |
| Detenido por el usuario | ■ Stopped. | ■ Detenido. |
| Primera compilación lenta | Preparing Go for the first time. This takes about a minute and only happens once. | Preparando Go por primera vez. Tarda cerca de un minuto y solo pasa una vez. |
| Depurando | ● Debugging {file} · paused at line {line} | ● Depurando {file} · en pausa en la línea {line} |
| stdin al depurar | While debugging, your program can't read the keyboard. | Mientras depuras, tu programa no puede leer el teclado. |
| Iniciando depurador | Starting the debugger… | Iniciando el depurador… |

### Errores del IDE (qué pasó + por qué + cómo arreglarlo)
| Situación | English | Español | Botones |
|---|---|---|---|
| Go no encontrado | Go isn't installed. VizcachaIDE needs Go to run your programs. Install Go or choose where it is in Settings. | Go no está instalado. VizcachaIDE necesita Go para ejecutar tus programas. Instálalo o indica dónde está en Configuración. | Install Go · Choose location / Instalar Go · Elegir ubicación |
| Delve no encontrado | The debugger (Delve) isn't available. Reinstall VizcachaIDE or install Delve with the command below. | El depurador (Delve) no está disponible. Reinstala VizcachaIDE o instala Delve con el comando de abajo. | Copy command / Copiar comando |
| gopls no encontrado | Smart suggestions are off because gopls isn't available. Basic suggestions still work. | Las sugerencias inteligentes están apagadas porque gopls no está disponible. Las sugerencias básicas siguen funcionando. | — |
| No se pudo guardar | Couldn't save {file}. {reason} Check that the folder exists and you can write to it. | No se pudo guardar {file}. {reason} Revisa que la carpeta exista y que puedas escribir en ella. | Try again / Reintentar |
| gofmt rechaza el código | Not formatted: Go couldn't read line {line}. Your file was saved as it is. | No se formateó: Go no pudo leer la línea {line}. Tu archivo se guardó tal cual. | — |
| Argumentos inválidos | The program arguments have an unclosed quote. | Los argumentos del programa tienen una comilla sin cerrar. | — |

### Confirmaciones (la acción en el botón)
| Situación | English | Español | Botones EN / ES |
|---|---|---|---|
| Cerrar con cambios | Save changes to {file} before closing? | ¿Guardar los cambios de {file} antes de cerrar? | Save · Don't save · Cancel / Guardar · No guardar · Cancelar |
| Archivo cambiado fuera | {file} changed outside VizcachaIDE. Reload it and lose your unsaved changes? | {file} cambió fuera de VizcachaIDE. ¿Recargarlo y perder tus cambios sin guardar? | Reload · Keep mine / Recargar · Conservar los míos |
| Reemplazar todo | Replace {count} matches in {file}? | ¿Reemplazar {count} coincidencias en {file}? | Replace {count} · Cancel / Reemplazar {count} · Cancelar |

### Primer arranque (una idea por paso)
| Paso | English | Español |
|---|---|---|
| Bienvenida | Welcome to VizcachaIDE. Let's get you writing Go in under a minute. | Te damos la bienvenida a VizcachaIDE. En menos de un minuto estarás escribiendo Go. |
| 1 de 3 | Choose your language | Elige tu idioma |
| 2 de 3 | Checking Go… Go {version} is ready. | Revisando Go… Go {version} está listo. |
| 3 de 3 | Open your first program | Abre tu primer programa |
| Botones paso 3 | Open hello example · Start with a blank file | Abrir el ejemplo «Hola» · Empezar con un archivo en blanco |

---

## Alternatives (decisiones con más de una opción razonable)

| Elemento | Opción | Copy EN / ES | Tono | Mejor para |
|---|---|---|---|---|
| Paso sobre | **A (elegida)** | Next line / Siguiente línea | Llano | Principiantes: describe lo que ven pasar |
| | B | Step over / Paso sobre | Técnico | Usuarios de otros IDE (queda en la ayuda emergente) |
| | C | Run this line / Ejecutar esta línea | Llano | Ambiguo si la línea llama funciones |
| Pila de llamadas | **A (elegida)** | How you got here / Cómo llegaste aquí | Didáctico | Explica para qué sirve el panel |
| | B | Call stack / Pila de llamadas | Técnico | Modo experto futuro |
| Panel de errores | **A (elegida)** | Assistant / Asistente | Neutro | Continuidad con la 1.0 y la documentación |
| | B | Explain / Explicación | Directo | Si el panel solo explicara errores (también muestra variables) |

## Rationale

- **El usuario está aprendiendo y suele estar frustrado** cuando aparece un error. Por eso los errores empiezan por un hecho concreto ("La variable «resultado» nunca se usa") y no por "Error". Además, siempre llevan una acción.
- **Los botones dicen lo que hacen** ("Guardar", "Recargar", "Reemplazar 12"). Nada de "Aceptar" ni "Sí".
- **Términos técnicos de Go** (breakpoint, goroutine, gopls, Delve): se conservan donde el usuario los va a encontrar fuera del IDE, como "goroutine" o los errores de Go. En las acciones se usa el nombre llano y el técnico va en la ayuda emergente.
- **"Punto de interrupción"** en todo el producto en español; nunca "breakpoint".

## Localization Notes

- **Placeholders con nombre** (`{file}`, `{line}`, `{count}`): no traducirlos ni reordenarlos fuera de la frase.
- **Plurales** con ICU MessageFormat (`{count, plural, one {…} other {…}}`) en svelte-i18n y con go-i18n en el backend.
- **El español es entre un 20 y un 30 % más largo.** Los botones de la barra deben admitir hasta 22 caracteres ("Terminar depuración") sin cortarse. Probar a 1024 px de ancho.
- **Tuteo** en español neutro latinoamericano ("Pulsa", "Abre"). Evitar el voseo y los regionalismos.
- **Evitar el género gramatical** cuando el texto se refiere a la persona: "Te damos la bienvenida", no "Bienvenido/a".
- **Comillas:** en español, «angulares» para nombres de variables en la prosa; en inglés, “curly quotes”.
- **Decimales:** "0,4 s" en español y "0.4 s" en inglés. Formatear con `Intl.NumberFormat`, nunca a mano.
- **Atajos de teclado** (F5, Ctrl+F10): no se traducen. En macOS, Wails/CodeMirror los muestran con ⌘.
