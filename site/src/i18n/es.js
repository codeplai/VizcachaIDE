// Copy en español (latinoamericano neutro, tuteo). Tono: un profesor paciente.
export default {
  htmlLang: 'es-PE',
  skipToContent: 'Saltar al contenido',
  langLabel: 'Idioma',
  themeLabel: 'Cambiar entre tema claro y oscuro',
  menuLabel: 'Abrir o cerrar el menú',
  navLabel: 'Navegación principal',
  nav: { features: 'Qué puedes hacer', download: 'Descargar', manual: 'Manual', source: 'Código' },
  headerCta: 'Descargar',
  parentLink: 'Codeplai',

  meta: {
    title: 'VizcachaIDE: el IDE de Go para quien recién empieza',
    description:
      'Escribe, ejecuta, entiende tus errores y depura paso a paso en Go, en español o en inglés. Gratis, de código abierto y hecho en Perú.',
    manualTitle: 'Manual de VizcachaIDE: instalar, ejecutar y depurar',
    manualDescription:
      'Guía breve para instalar VizcachaIDE, ejecutar tu primer programa, entender los errores, depurar paso a paso, atajos de teclado y solución de problemas.',
  },

  hero: {
    badge: 'Versión candidata 2.0.0-rc1 · Gratis y de código abierto',
    h1: 'El IDE de Go para quien recién empieza',
    sub: 'Escribe tu programa, ejecútalo con un botón, entiende por qué falla y míralo funcionar paso a paso. Todo en una sola ventana, en español o en inglés.',
    ctaPrimary: 'Descargar',
    ctaSecondary: 'Ver el código',
    note: 'Inspirado en Thonny, el IDE con el que miles de personas aprenden Python.',
    logoAlt: 'Logo de VizcachaIDE by Codeplai: una vizcacha con audífonos y lentes con código',
    shotAlt:
      'VizcachaIDE ejecutando un programa en Go: el editor arriba y la salida del programa abajo',
    shotCaption: 'Pulsa Ejecutar (F5) y mira la salida de tu programa al instante.',
  },

  audience: {
    eyebrow: 'Para quién es',
    title: 'Para quien quiere aprender Go sin pelearse con la configuración',
    items: [
      {
        t: 'Estudiantes',
        d: 'Que dan sus primeros pasos en programación con Go.',
      },
      {
        t: 'Docentes',
        d: 'Que necesitan una herramienta sencilla para el aula, que funcione en español y sin configuraciones complicadas.',
      },
      {
        t: 'Cualquier persona',
        d: 'Que quiera aprender Go sin pelearse primero con terminales, variables de entorno y extensiones.',
      },
    ],
  },

  features: {
    eyebrow: 'Qué puedes hacer',
    title: 'Todo lo que necesitas para aprender, en una sola ventana',
    run: {
      t: 'Escribir y ejecutar',
      items: [
        'Pulsa **Ejecutar (F5)** y mira la salida de tu programa al instante.',
        'Escribe en la consola cuando tu programa te pida datos por teclado.',
        'Pasa argumentos a tu programa y trabaja con proyectos que usan `go.mod`.',
        'Tu código se ordena solo al guardar, con el formato estándar de Go (gofmt).',
      ],
    },
    errors: {
      t: 'Entender tus errores',
      items: [
        'Cuando Go encuentra un problema, el **Asistente** te explica **qué pasó y cómo arreglarlo**, en tu idioma.',
        'Reconoce 25 de los errores más comunes de quien empieza: variables sin usar, imports de más, tipos que no encajan, índices fuera de rango, mapas sin inicializar, bloqueos entre goroutines y más.',
        'El mensaje original de Go siempre está a la vista, con un botón para buscarlo en internet. Así aprendes a leer los errores reales.',
        'Los errores se subrayan **mientras escribes**, antes de ejecutar.',
      ],
      alt: 'El Asistente de VizcachaIDE explicando un error de Go en español',
      caption: 'El Asistente explica qué pasó y cómo arreglarlo.',
    },
    debug: {
      t: 'Ver tu programa paso a paso',
      items: [
        'Haz clic junto a un número de línea para poner un **punto de interrupción** y pulsa **Depurar (F6)**.',
        'Avanza con botones que hablan claro: **Siguiente línea**, **Entrar en la función**, **Salir de la función**.',
        'Mira el valor de tus variables en cada paso. La que acaba de cambiar se resalta, para que veas qué hizo la última línea.',
        'Descubre **cómo llegaste ahí** (la pila de llamadas) y qué hace cada goroutine.',
      ],
      alt: 'El depurador de VizcachaIDE mostrando las variables, la que acaba de cambiar y la pila de llamadas',
      caption: 'Depurando paso a paso: variables, la que acaba de cambiar y cómo llegaste ahí.',
    },
    fast: {
      t: 'Escribir más rápido',
      items: [
        'Autocompletado inteligente de Go, con la documentación de cada función.',
        'Ayuda con los parámetros mientras escribes una llamada.',
        'Ctrl+clic para ir a donde se define una función.',
        'Buscar y reemplazar, ir a una línea, zoom, tema claro u oscuro.',
      ],
    },
  },

  learning: {
    eyebrow: 'Pensado para aprender',
    title: 'Menos botones, más claridad',
    items: [
      {
        t: 'Español e inglés',
        d: 'En toda la interfaz y en las explicaciones de errores. Detecta el idioma de tu sistema y puedes cambiarlo cuando quieras.',
      },
      {
        t: 'Un botón principal',
        d: 'Ejecutar es lo más visible de la ventana; lo demás aparece cuando lo necesitas.',
      },
      {
        t: 'Letra muy legible',
        d: 'Usa Atkinson Hyperlegible, una tipografía diseñada para que caracteres como 0 y O, o 1, l e I, no se confundan.',
      },
      {
        t: 'Todo incluido',
        d: 'La versión completa trae Go, el depurador Delve y gopls: instalas VizcachaIDE y ya puedes programar.',
      },
    ],
  },

  download: {
    eyebrow: 'Descarga',
    title: 'Descarga VizcachaIDE',
    lead: 'Es gratis y de código abierto (licencia MIT). Elige la versión completa si no tienes Go instalado.',
    version: 'Versión candidata (release candidate) 2.0.0-rc1',
    tableCaption: 'Versiones disponibles de VizcachaIDE para Windows',
    cols: ['Versión', 'Qué incluye', 'Para quién', 'Archivos en Windows'],
    rows: [
      {
        name: 'Completa',
        tag: 'Recomendada para empezar',
        includes: 'VizcachaIDE + Go 1.25 + Delve + gopls',
        who: 'Si no tienes Go instalado.',
        files: [
          'Instalador (setup.exe), unos 57 MB. Por usuario, sin permisos de administrador.',
          'Portable (.zip), unos 87 MB. No necesita instalación.',
        ],
      },
      {
        name: 'Ligera',
        tag: null,
        includes: 'Solo VizcachaIDE',
        who: 'Si ya tienes Go instalado.',
        files: ['Unos 6 MB. Necesita Go en tu sistema.'],
      },
    ],
    systemsTitle: 'Sistemas',
    systems: [
      '**Windows 10 y 11 (x64):** instalador sin permisos de administrador y versión portable.',
      '**macOS y Linux:** versión preliminar, todavía sin probar en equipos reales.',
    ],
    webview:
      'La versión portable no necesita instalar nada más, salvo WebView2, que ya viene en Windows 11 y en casi todos los Windows 10 actualizados.',
    button: 'Ir a las descargas en GitHub',
    buttonNote: 'Se abre la página de versiones del proyecto. Descarga el archivo que corresponda a tu elección.',
    manualLink: 'Cómo instalarlo y usarlo: el manual',
  },

  safety: {
    id: 'un-proyecto-nuevo',
    eyebrow: 'Un proyecto nuevo',
    title: 'Por qué Windows puede avisarte y por qué es seguro',
    intro: [
      'VizcachaIDE es un **proyecto nuevo e independiente**, hecho en Perú por [Codeplai Games](https://codeplai.pe). Sus instaladores **todavía no tienen firma digital**, así que la primera vez que abras uno, SmartScreen de Windows puede mostrar *«Windows protegió tu PC»* y decir que el editor es desconocido.',
      'Este aviso aparece con todo programa nuevo que no está firmado; no significa que se haya encontrado un virus.',
    ],
    stepsTitle: 'Para continuar',
    steps: [
      'En el aviso «Windows protegió tu PC», haz clic en **Más información**.',
      'Haz clic en **Ejecutar de todas formas**.',
    ],
    macNote: 'En macOS la app aún no está notarizada: haz clic derecho sobre ella y elige **Abrir**.',
    proofTitle: 'No tienes que creernos sin más',
    proofs: [
      {
        t: 'El código fuente es público',
        d: 'Cada línea está en [github.com/codeplai/VizcachaIDE](https://github.com/codeplai/VizcachaIDE), con licencia MIT. Puedes leerlo y también compilarlo tú mismo.',
      },
      {
        t: 'Puedes verificar tu descarga',
        d: 'Cada versión publicada incluye un archivo `SHA256SUMS`. En PowerShell, `Get-FileHash .\\VizcachaIDE-…zip` debe mostrar el mismo valor que aparece en ese archivo.',
      },
      {
        t: 'Las herramientas incluidas son las oficiales',
        d: 'Go, Delve y gopls se descargan de sus fuentes oficiales, y el script de empaquetado comprueba cada descarga con un SHA-256 fijo.',
      },
    ],
    signature:
      '**Sobre la firma:** planeamos adquirir un certificado de firma de código a medida que el proyecto crezca, para que Windows reconozca al editor y el aviso desaparezca. Mientras tanto, el código público y las sumas de verificación son la forma de comprobar lo que instalas.',
  },

  status: {
    eyebrow: 'Estado del proyecto',
    title: 'Versión candidata: ya se puede usar',
    lead: 'VizcachaIDE está en **versión candidata (release candidate)**: ya se puede usar y estamos puliendo detalles antes de la versión final.',
    items: [
      {
        t: 'Nueva edición 2.0',
        d: 'Interfaz rediseñada, más ligera y rápida, y con un depurador que explica cada paso.',
      },
      {
        t: 'Edición clásica 1.x',
        d: 'Sigue disponible mientras la 2.0 llega a su versión final.',
      },
    ],
  },

  peru: {
    eyebrow: 'Hecho en Perú',
    title: 'Un proyecto de Codeplai Games',
    p1: 'VizcachaIDE es un proyecto de [Codeplai Games](https://codeplai.pe), creado por Marks Calderon, CEO de Codeplai. Su nombre viene de la **vizcacha**, el roedor de los Andes que vive entre las rocas de la sierra: pequeño, curioso y siempre atento.',
    contactTitle: 'Contacto',
    p2: '¿Tienes ideas, encontraste un error o quieres usarlo en tu clase? Escríbenos a [hola@codeplai.pe](mailto:hola@codeplai.pe) o abre un *issue* en GitHub.',
    issue: 'Abrir un issue',
    visit: 'Conoce Codeplai',
  },

  footer: {
    line: 'El IDE de Go para quien recién empieza. Gratis y de código abierto (MIT).',
    navTitle: 'Proyecto',
    moreTitle: 'Contacto',
    parent: 'Un proyecto de Codeplai Games',
    license:
      'VizcachaIDE se distribuye bajo licencia MIT e incluye Go, Delve y gopls, cada uno con su propia licencia de código abierto.',
  },

  manual: {
    eyebrow: 'Manual',
    title: 'Manual de VizcachaIDE',
    lead: 'Una guía breve para instalar, escribir tu primer programa, entender los errores y depurar. Si algo no queda claro, escríbenos a hola@codeplai.pe.',
    tocTitle: 'En esta página',
    sections: [
      {
        id: 'instalar',
        t: 'Instalar',
        blocks: [
          {
            p: 'Descarga el instalador desde la [página de versiones](https://github.com/codeplai/VizcachaIDE/releases). Hay dos variantes:',
          },
          {
            ul: [
              '**Completa:** VizcachaIDE + Go 1.25 + Delve + gopls. Es la que debes elegir si no tienes Go instalado.',
              '**Ligera:** solo VizcachaIDE (unos 6 MB). Elígela si ya tienes Go instalado; para depurar necesitas además `dlv` (Delve) y, para el autocompletado, `gopls`.',
            ],
          },
          { h: 'Windows 10 y 11 (x64)' },
          {
            ul: [
              '**Instalador (setup.exe):** se instala solo para tu usuario, sin permisos de administrador.',
              '**Portable (.zip):** descomprímelo en una carpeta y abre VizcachaIDE. No necesita instalación ni DLLs adicionales, solo WebView2 (ya viene en Windows 11 y en casi todos los Windows 10 actualizados).',
            ],
          },
          {
            p: 'Como el instalador todavía no tiene firma digital, Windows puede mostrar *«Windows protegió tu PC»*. Haz clic en **Más información → Ejecutar de todas formas**. [Aquí explicamos por qué es seguro](/#un-proyecto-nuevo).',
          },
          { h: 'macOS y Linux' },
          {
            p: 'Son versiones preliminares, aún sin probar en equipos reales. En macOS la app no está notarizada: haz clic derecho sobre ella y elige **Abrir**. En Linux, da permiso de ejecución al AppImage (`chmod +x VizcachaIDE-*.AppImage`) y ejecútalo.',
          },
          { h: 'Verificar la descarga' },
          {
            p: 'Cada versión incluye un archivo `SHA256SUMS`. En PowerShell, ejecuta `Get-FileHash .\\VizcachaIDE-…zip`: el valor debe ser igual al del archivo.',
          },
        ],
      },
      {
        id: 'primer-programa',
        t: 'Tu primera vez',
        blocks: [
          {
            p: 'Al abrir VizcachaIDE verás una pestaña vacía lista para escribir. La interfaz usa el idioma de tu sistema (español o inglés).',
          },
          {
            ol: [
              'Escribe un programa, por ejemplo el siguiente.',
              'Pulsa **F5** para ejecutarlo. La salida aparece en el panel inferior.',
              'Guarda el archivo con **Ctrl+S**.',
            ],
          },
          {
            code: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("¡Hola, VizcachaIDE!")\n}',
          },
        ],
      },
      {
        id: 'ejecutar',
        t: 'Ejecutar tu programa',
        blocks: [
          {
            ul: [
              '**Ejecutar (F5):** compila y ejecuta tu programa. **Detener (Shift+F5)** lo termina cuando quieras.',
              'Si tu programa pide datos por teclado, escribe en la consola y pulsa Enter.',
              'Puedes pasar argumentos a tu programa y abrir proyectos que usan `go.mod`.',
              'Al guardar, tu código se ordena con el formato estándar de Go (gofmt).',
            ],
          },
        ],
      },
      {
        id: 'errores',
        t: 'Entender los errores',
        blocks: [
          {
            p: 'Cuando Go encuentra un problema, el **Asistente** (panel de la derecha) te explica qué pasó y cómo arreglarlo, en español o en inglés. Reconoce 25 errores comunes, como una variable declarada y no usada, un import de más, tipos que no encajan, un índice fuera de rango o un mapa sin inicializar.',
          },
          {
            ul: [
              'Los errores se subrayan mientras escribes, antes de ejecutar.',
              'El mensaje original de Go siempre se muestra, con un botón para buscarlo en internet. Así aprendes a leer los errores reales.',
            ],
          },
          {
            p: 'Un buen ejercicio: comete un error a propósito (declara una variable que nunca uses), pulsa F5 y lee la explicación del Asistente.',
          },
        ],
      },
      {
        id: 'depurar',
        t: 'Depurar paso a paso',
        blocks: [
          {
            ol: [
              'Guarda tu archivo (la depuración necesita un archivo guardado).',
              'Haz clic junto a un número de línea para poner un **punto de interrupción**.',
              'Pulsa **Depurar (F6)**. El programa se detiene en esa línea.',
              'Avanza con **Siguiente línea (F7)**, **Entrar en la función (F8)** o **Salir de la función (F9)**.',
              'Mira tus variables en cada paso: la que acaba de cambiar se resalta. El panel «Cómo llegaste aquí» muestra la pila de llamadas.',
              '**Continuar (Shift+F6)** sigue hasta el próximo punto de interrupción; **Terminar depuración (Shift+F5)** termina la sesión.',
            ],
          },
          {
            p: 'Con **Ejecutar hasta aquí (Ctrl+F10)** el programa avanza hasta la línea donde está el cursor. La depuración usa Delve, que ya viene en la versión completa.',
          },
        ],
      },
      {
        id: 'atajos',
        t: 'Atajos de teclado',
        blocks: [
          {
            table: [
              ['Ejecutar / Detener', 'F5 / Shift+F5'],
              ['Depurar / Continuar', 'F6 / Shift+F6'],
              ['Siguiente línea / Entrar / Salir', 'F7 / F8 / F9'],
              ['Ejecutar hasta aquí', 'Ctrl+F10'],
              ['Guardar', 'Ctrl+S'],
              ['Buscar', 'Ctrl+F'],
              ['Ir a una línea', 'Ctrl+G'],
              ['Ir a la definición', 'F12 o Ctrl+clic'],
              ['Autocompletar', 'Ctrl+Space'],
              ['Zoom: acercar / alejar / restablecer', 'Ctrl + / Ctrl − / Ctrl 0'],
            ],
            head: ['Acción', 'Atajo'],
            note: 'En macOS, usa ⌘ en lugar de Ctrl.',
          },
        ],
      },
      {
        id: 'problemas',
        t: 'Solución de problemas',
        blocks: [
          { h: 'Windows muestra «Windows protegió tu PC»' },
          {
            p: 'Es normal con programas nuevos sin firma digital. Haz clic en **Más información → Ejecutar de todas formas**. Puedes comprobar tu descarga con `SHA256SUMS` y revisar el [código fuente](https://github.com/codeplai/VizcachaIDE).',
          },
          { h: 'La ventana no abre o aparece en blanco (WebView2)' },
          {
            p: 'VizcachaIDE usa WebView2, que viene incluido en Windows 11 y en casi todos los Windows 10 actualizados. Si tu equipo no lo tiene, instala el «WebView2 Runtime» desde la página de Microsoft y vuelve a abrir VizcachaIDE. Actualizar Windows también suele resolverlo.',
          },
          { h: 'Dice que no encuentra Go' },
          {
            p: 'Si usas la versión **ligera**, necesitas tener Go instalado y disponible en el `PATH`. Si no quieres instalarlo, usa la versión **completa**, que ya lo incluye.',
          },
          { h: 'El depurador o el autocompletado no funcionan' },
          {
            p: 'En la versión ligera necesitas instalar Delve y gopls:',
          },
          {
            code: 'go install github.com/go-delve/delve/cmd/dlv@latest\ngo install golang.org/x/tools/gopls@latest',
          },
          { h: 'Encontraste un error o tienes una idea' },
          {
            p: 'Abre un [issue en GitHub](https://github.com/codeplai/VizcachaIDE/issues) o escríbenos a [hola@codeplai.pe](mailto:hola@codeplai.pe).',
          },
        ],
      },
    ],
  },
};
