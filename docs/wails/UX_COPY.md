# UX Copy: VizcachaIDE Wails (EN / ES)

Source of truth for the interface texts of the Wails variant. The coders use these texts **verbatim**. Any new text follows the same rules:
- **Clear:** no jargon, or with the technical term only in the tooltip.
- **Brief.**
- **Consistent:** one concept, one word.
- **Useful:** every error says what happened, why, and how to fix it.
- **Human.**

**Tone:** a patient teacher. Calm when something fails, brief when everything goes well, and never condescending.

The Spanish column is the product's Spanish copy and stays in Spanish; everything else in this document is in English.

---

## Recommended Copy

### Main actions
| Element | English | Español | Tooltip (EN / ES) |
|---|---|---|---|
| Run | **Run** | **Ejecutar** | Run your program (F5) / Ejecuta tu programa (F5) |
| Debug | **Debug** | **Depurar** | Run step by step and see your variables (F6) / Ejecuta paso a paso y mira tus variables (F6) |
| End debugging | **Stop debugging** | **Terminar depuración** | Shift+F5 |
| Stop program | **Stop** | **Detener** | Stop the running program (Shift+F5) / Detén el programa en ejecución (Shift+F5) |
| Step over | **Next line** | **Siguiente línea** | Runs this line without going inside functions (Step over) / Ejecuta esta línea sin entrar en funciones (Step over) |
| Step into | **Go into function** | **Entrar en la función** | (Step into) |
| Step out | **Leave function** | **Salir de la función** | (Step out) |
| Continue | **Continue** | **Continuar** | Run until the next breakpoint (Shift+F6) / Ejecuta hasta el próximo punto de interrupción (Shift+F6) |
| Run to cursor | **Run to here** | **Ejecutar hasta aquí** | Run until the line with the cursor (Ctrl+F10) / Ejecuta hasta la línea del cursor (Ctrl+F10) |

### Panels
| Element | English | Español |
|---|---|---|
| Right panel | Assistant | Asistente |
| Output | Output | Salida |
| Problems | Problems | Problemas |
| Call stack | How you got here | Cómo llegaste aquí |
| Variables | Variables in {function}() | Variables en {function}() |
| Just-changed variable | just changed | acaba de cambiar |
| Outline | Outline | Estructura |
| Files | Files | Archivos |
| Settings | Settings | Configuración |

### Empty states (what it is + why it is empty + how to start)
| Place | English | Español | CTA |
|---|---|---|---|
| Files | No folder open. Open a folder to see your project's files here. | No hay ninguna carpeta abierta. Abre una para ver aquí los archivos de tu proyecto. | Open folder / Abrir carpeta |
| Output | Nothing here yet. Press Run (F5) to see what your program prints. | Todavía no hay nada. Pulsa Ejecutar (F5) para ver lo que imprime tu programa. | — |
| Problems | No problems found. Go checks your code as you type. | No hay problemas. Go revisa tu código mientras escribes. | — |
| Variables | Variables appear when the program pauses. Add a breakpoint and press Debug. | Las variables aparecen cuando el programa se pausa. Pon un punto de interrupción y pulsa Depurar. | — |
| Outline | Functions and types of this file appear here once Go has read it. | Las funciones y tipos de este archivo aparecen aquí cuando Go termina de leerlo. | — |
| Assistant (tips) | **Press Run or F5** to run your program. · **Something wrong?** You'll see what happened and how to fix it here, in your language. · **To watch your program step by step**, click next to a line number and press Debug. | **Pulsa Ejecutar o F5** para correr tu programa. · **¿Algo falla?** Aquí verás qué pasó y cómo arreglarlo, en tu idioma. · **Para ver tu programa paso a paso**, haz clic junto a un número de línea y pulsa Depurar. | — |

### Run messages
| Situation | English | Español |
|---|---|---|
| Start | ▶ Running {file}… | ▶ Ejecutando {file}… |
| Success | ✓ Finished in {seconds} s | ✓ Terminó bien en {seconds} s |
| Exit code ≠ 0 | ✗ Your program ended with code {code}. | ✗ Tu programa terminó con código {code}. |
| Compile failure | ✗ Your program didn't run: Go found {count, plural, one {# problem} other {# problems}}. See the explanation on the right. | ✗ No se pudo ejecutar: Go encontró {count, plural, one {# problema} other {# problemas}}. Mira la explicación a la derecha. |
| Stopped by the user | ■ Stopped. | ■ Detenido. |
| Slow first compilation | Preparing Go for the first time. This takes about a minute and only happens once. | Preparando Go por primera vez. Tarda cerca de un minuto y solo pasa una vez. |
| Debugging | ● Debugging {file} · paused at line {line} | ● Depurando {file} · en pausa en la línea {line} |
| stdin while debugging | While debugging, your program can't read the keyboard. | Mientras depuras, tu programa no puede leer el teclado. |
| Starting the debugger | Starting the debugger… | Iniciando el depurador… |

### IDE errors (what happened + why + how to fix it)
| Situation | English | Español | Buttons |
|---|---|---|---|
| Go not found | Go isn't installed. VizcachaIDE needs Go to run your programs. Install Go or choose where it is in Settings. | Go no está instalado. VizcachaIDE necesita Go para ejecutar tus programas. Instálalo o indica dónde está en Configuración. | Install Go · Choose location / Instalar Go · Elegir ubicación |
| Delve not found | The debugger (Delve) isn't available. Reinstall VizcachaIDE or install Delve with the command below. | El depurador (Delve) no está disponible. Reinstala VizcachaIDE o instala Delve con el comando de abajo. | Copy command / Copiar comando |
| gopls not found | Smart suggestions are off because gopls isn't available. Basic suggestions still work. | Las sugerencias inteligentes están apagadas porque gopls no está disponible. Las sugerencias básicas siguen funcionando. | — |
| Could not save | Couldn't save {file}. {reason} Check that the folder exists and you can write to it. | No se pudo guardar {file}. {reason} Revisa que la carpeta exista y que puedas escribir en ella. | Try again / Reintentar |
| gofmt rejects the code | Not formatted: Go couldn't read line {line}. Your file was saved as it is. | No se formateó: Go no pudo leer la línea {line}. Tu archivo se guardó tal cual. | — |
| Invalid arguments | The program arguments have an unclosed quote. | Los argumentos del programa tienen una comilla sin cerrar. | — |

### Confirmations (the action on the button)
| Situation | English | Español | Buttons EN / ES |
|---|---|---|---|
| Close with changes | Save changes to {file} before closing? | ¿Guardar los cambios de {file} antes de cerrar? | Save · Don't save · Cancel / Guardar · No guardar · Cancelar |
| File changed externally | {file} changed outside VizcachaIDE. Reload it and lose your unsaved changes? | {file} cambió fuera de VizcachaIDE. ¿Recargarlo y perder tus cambios sin guardar? | Reload · Keep mine / Recargar · Conservar los míos |
| Replace all | Replace {count} matches in {file}? | ¿Reemplazar {count} coincidencias en {file}? | Replace {count} · Cancel / Reemplazar {count} · Cancelar |

### First run (one idea per step)
| Step | English | Español |
|---|---|---|
| Welcome | Welcome to VizcachaIDE. Let's get you writing Go in under a minute. | Te damos la bienvenida a VizcachaIDE. En menos de un minuto estarás escribiendo Go. |
| 1 of 3 | Choose your language | Elige tu idioma |
| 2 of 3 | Checking Go… Go {version} is ready. | Revisando Go… Go {version} está listo. |
| 3 of 3 | Open your first program | Abre tu primer programa |
| Step 3 buttons | Open hello example · Start with a blank file | Abrir el ejemplo «Hola» · Empezar con un archivo en blanco |

---

## Alternatives (decisions with more than one reasonable option)

| Element | Option | Copy EN / ES | Tone | Best for |
|---|---|---|---|---|
| Step over | **A (chosen)** | Next line / Siguiente línea | Plain | Beginners: describes what they see happen |
| | B | Step over / Paso sobre | Technical | Users of other IDEs (kept in the tooltip) |
| | C | Run this line / Ejecutar esta línea | Plain | Ambiguous if the line calls functions |
| Call stack | **A (chosen)** | How you got here / Cómo llegaste aquí | Didactic | Explains what the panel is for |
| | B | Call stack / Pila de llamadas | Technical | Future expert mode |
| Error panel | **A (chosen)** | Assistant / Asistente | Neutral | Continuity with 1.0 and the documentation |
| | B | Explain / Explicación | Direct | If the panel only explained errors (it also shows variables) |

## Rationale

- **The user is learning and is usually frustrated** when an error appears. That is why errors start with a concrete fact (ES: "La variable «resultado» nunca se usa", i.e. "The variable «resultado» is never used") rather than with "Error". They also always include an action.
- **Buttons say what they do** ("Guardar", "Recargar", "Reemplazar 12" in Spanish; "Save", "Reload", "Replace 12" in English). No "OK" or "Yes".
- **Go technical terms** (breakpoint, goroutine, gopls, Delve): kept where the user will meet them outside the IDE, such as "goroutine" or Go's own errors. In actions the plain name is used and the technical one goes in the tooltip.
- **"Punto de interrupción"** throughout the Spanish product; never "breakpoint".

## Localization Notes

- **Named placeholders** (`{file}`, `{line}`, `{count}`): do not translate them or move them out of the sentence.
- **Plurals** with ICU MessageFormat (`{count, plural, one {…} other {…}}`) in svelte-i18n and with go-i18n in the backend.
- **Spanish is 20 to 30% longer.** Toolbar buttons must fit up to 22 characters ("Terminar depuración") without being cut off. Test at 1024 px width.
- **Informal "you" (tuteo)** in neutral Latin American Spanish ("Pulsa", "Abre"). Avoid voseo and regionalisms.
- **Avoid grammatical gender** when the text refers to the person: "Te damos la bienvenida", not "Bienvenido/a".
- **Quotation marks:** in Spanish, «angle quotes» for variable names in prose; in English, “curly quotes”.
- **Decimals:** "0,4 s" in Spanish and "0.4 s" in English. Format with `Intl.NumberFormat`, never by hand.
- **Keyboard shortcuts** (F5, Ctrl+F10): not translated. On macOS, Wails/CodeMirror show them with ⌘.
