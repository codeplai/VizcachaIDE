// English copy. Tone: a patient teacher.
export default {
  htmlLang: 'en',
  skipToContent: 'Skip to content',
  langLabel: 'Language',
  themeLabel: 'Switch between light and dark theme',
  menuLabel: 'Open or close the menu',
  navLabel: 'Main navigation',
  nav: { features: 'What you can do', download: 'Download', manual: 'Manual', source: 'Source' },
  headerCta: 'Download',
  parentLink: 'Codeplai',

  meta: {
    title: 'VizcachaIDE: the Go IDE for people just starting out',
    description:
      'Write, run, understand your errors and debug step by step in Go, in English or Spanish. Free, open source and made in Peru.',
    manualTitle: 'VizcachaIDE manual: install, run and debug',
    manualDescription:
      'A short guide to install VizcachaIDE, run your first program, understand errors, debug step by step, keyboard shortcuts and troubleshooting.',
  },

  hero: {
    badge: 'Release candidate 2.0.0-rc1 · Free and open source',
    h1: 'The Go IDE for people just starting out',
    sub: 'Write your program, run it with one button, understand why it fails and watch it work step by step. All in a single window, in English or Spanish.',
    ctaPrimary: 'Download',
    ctaSecondary: 'View the source',
    note: 'Inspired by Thonny, the IDE thousands of people use to learn Python.',
    logoAlt: 'VizcachaIDE by Codeplai logo: a vizcacha wearing headphones and code-display goggles',
    shotAlt: 'VizcachaIDE running a Go program: the editor on top and the program output below',
    shotCaption: 'Press Run (F5) and see your program output right away.',
  },

  audience: {
    eyebrow: 'Who it is for',
    title: 'For anyone who wants to learn Go without fighting the setup',
    items: [
      { t: 'Students', d: 'Taking their first steps in programming with Go.' },
      {
        t: 'Teachers',
        d: 'Who need a simple tool for the classroom, available in Spanish and English and with no complicated setup.',
      },
      {
        t: 'Anyone',
        d: 'Who wants to learn Go without first fighting terminals, environment variables and extensions.',
      },
    ],
  },

  features: {
    eyebrow: 'What you can do',
    title: 'Everything you need to learn, in a single window',
    run: {
      t: 'Write and run',
      items: [
        'Press **Run (F5)** and see your program output right away.',
        'Type in the console when your program asks for keyboard input.',
        'Pass arguments to your program and work with projects that use `go.mod`.',
        'Your code is tidied up when you save, using the standard Go format (gofmt).',
      ],
    },
    errors: {
      t: 'Understand your errors',
      items: [
        'When Go finds a problem, the **Assistant** explains **what happened and how to fix it**, in your language.',
        'It recognizes 25 of the most common beginner errors: unused variables, extra imports, mismatched types, index out of range, uninitialized maps, goroutine deadlocks and more.',
        "Go's original message is always visible, with a button to search for it online. That way you learn to read the real errors.",
        'Errors are underlined **as you type**, before you run.',
      ],
      alt: 'The VizcachaIDE Assistant explaining a Go error in English',
      caption: 'The Assistant explains what happened and how to fix it.',
    },
    debug: {
      t: 'Watch your program step by step',
      items: [
        'Click next to a line number to set a **breakpoint**, then press **Debug (F6)**.',
        'Move forward with plainly named buttons: **Next line**, **Go into function**, **Leave function**.',
        'See the value of your variables at every step. The one that just changed is highlighted, so you can see what the last line did.',
        'Find out **how you got there** (the call stack) and what each goroutine is doing.',
      ],
      alt: 'The VizcachaIDE debugger showing variables, the one that just changed and the call stack',
      caption: 'Debugging step by step: variables, the one that just changed and how you got there.',
    },
    fast: {
      t: 'Write faster',
      items: [
        'Smart Go completion, with the documentation of each function.',
        'Parameter hints while you write a call.',
        'Ctrl+click to jump to where a function is defined.',
        'Find and replace, go to line, zoom, light or dark theme.',
      ],
    },
  },

  learning: {
    eyebrow: 'Designed for learning',
    title: 'Fewer buttons, more clarity',
    items: [
      {
        t: 'English and Spanish',
        d: 'Across the whole interface and in the error explanations. It detects your system language and you can change it any time.',
      },
      {
        t: 'One main button',
        d: 'Run is the most visible thing in the window; everything else shows up when you need it.',
      },
      {
        t: 'Very readable type',
        d: 'It uses Atkinson Hyperlegible, a typeface designed so characters like 0 and O, or 1, l and I, are not confused.',
      },
      {
        t: 'Everything included',
        d: 'The full version ships with Go, the Delve debugger and gopls: install VizcachaIDE and you can start coding.',
      },
    ],
  },

  download: {
    eyebrow: 'Download',
    title: 'Download VizcachaIDE',
    lead: 'It is free and open source (MIT license). Choose the full version if you do not have Go installed.',
    version: 'Release candidate 2.0.0-rc1',
    tableCaption: 'Available VizcachaIDE variants for Windows',
    cols: ['Variant', 'What it contains', 'Who it is for', 'Windows files'],
    rows: [
      {
        name: 'Full',
        tag: 'Recommended to start',
        includes: 'VizcachaIDE + Go 1.25 + Delve + gopls',
        who: 'If you do not have Go installed.',
        files: [
          'Installer (setup.exe), about 57 MB. Per user, no administrator rights.',
          'Portable (.zip), about 87 MB. No installation needed.',
        ],
      },
      {
        name: 'Lite',
        tag: null,
        includes: 'VizcachaIDE only',
        who: 'If you already have Go installed.',
        files: ['About 6 MB. Needs Go on your system.'],
      },
    ],
    systemsTitle: 'Systems',
    systems: [
      '**Windows 10 and 11 (x64):** installer with no administrator rights, and a portable version.',
      '**macOS and Linux:** preview, not yet tested on real machines.',
    ],
    webview:
      'The portable version needs nothing else installed except WebView2, which comes with Windows 11 and almost every updated Windows 10.',
    button: 'Go to the downloads on GitHub',
    buttonNote: "This opens the project's releases page. Download the file that matches your choice.",
    manualLink: 'How to install and use it: the manual',
  },

  safety: {
    id: 'a-new-project',
    eyebrow: 'A new project',
    title: 'Why Windows may warn you, and why it is safe',
    intro: [
      'VizcachaIDE is a **new, independent project**, made in Peru by [Codeplai Games](https://codeplai.pe). Its installers are **not digitally signed yet**, so the first time you open one, Windows SmartScreen may show *"Windows protected your PC"* and say the publisher is unknown.',
      'This warning appears for every new program without a signature. It does not mean that a virus was found.',
    ],
    stepsTitle: 'To continue',
    steps: [
      'On the "Windows protected your PC" warning, click **More info**.',
      'Click **Run anyway**.',
    ],
    macNote: 'On macOS the app is not notarized yet: right-click it and choose **Open**.',
    proofTitle: "You don't have to take our word for it",
    proofs: [
      {
        t: 'The source code is public',
        d: 'Every line is at [github.com/codeplai/VizcachaIDE](https://github.com/codeplai/VizcachaIDE), under the MIT license. You can read it, and you can build it yourself.',
      },
      {
        t: 'You can check your download',
        d: 'Each release includes a `SHA256SUMS` file. In PowerShell, `Get-FileHash .\\VizcachaIDE-…zip` must print the same value that is in that file.',
      },
      {
        t: 'The bundled tools are the official ones',
        d: 'Go, Delve and gopls are downloaded from their official sources, and the packaging script checks each download against a fixed SHA-256.',
      },
    ],
    signature:
      '**About the signature:** we plan to acquire a code-signing certificate as the project grows, so that Windows recognizes the publisher and the warning goes away. Until then, the public code and the checksums are how you can verify what you install.',
  },

  status: {
    eyebrow: 'Project status',
    title: 'Release candidate: ready to use',
    lead: 'VizcachaIDE is a **release candidate**: you can already use it, and we are polishing details before the final version.',
    items: [
      {
        t: 'New 2.0 edition',
        d: 'Redesigned interface, lighter and faster, with a debugger that explains every step.',
      },
      {
        t: 'Classic 1.x edition',
        d: 'Still available while 2.0 reaches its final version.',
      },
    ],
  },

  peru: {
    eyebrow: 'Made in Peru',
    title: 'A Codeplai Games project',
    p1: 'VizcachaIDE is a project by [Codeplai Games](https://codeplai.pe), created by Marks Calderon, CEO of Codeplai. Its name comes from the **vizcacha**, the Andean rodent that lives among the rocks of the highlands: small, curious and always alert.',
    contactTitle: 'Contact',
    p2: 'Have ideas, found a bug, or want to use it in your class? Write to us at [hola@codeplai.pe](mailto:hola@codeplai.pe) or open an issue on GitHub.',
    issue: 'Open an issue',
    visit: 'Visit Codeplai',
  },

  footer: {
    line: 'The Go IDE for people just starting out. Free and open source (MIT).',
    navTitle: 'Project',
    moreTitle: 'Contact',
    parent: 'A Codeplai Games project',
    license:
      'VizcachaIDE is distributed under the MIT license and includes Go, Delve and gopls, each with its own open source license.',
  },

  manual: {
    eyebrow: 'Manual',
    title: 'VizcachaIDE manual',
    lead: 'A short guide to install, write your first program, understand errors and debug. If something is unclear, write to hola@codeplai.pe.',
    tocTitle: 'On this page',
    sections: [
      {
        id: 'install',
        t: 'Install',
        blocks: [
          {
            p: 'Download the installer from the [releases page](https://github.com/codeplai/VizcachaIDE/releases). There are two variants:',
          },
          {
            ul: [
              '**Full:** VizcachaIDE + Go 1.25 + Delve + gopls. Choose it if you do not have Go installed.',
              '**Lite:** VizcachaIDE only (about 6 MB). Choose it if you already have Go installed; to debug you also need `dlv` (Delve) and, for completion, `gopls`.',
            ],
          },
          { h: 'Windows 10 and 11 (x64)' },
          {
            ul: [
              '**Installer (setup.exe):** installs for your user only, no administrator rights needed.',
              '**Portable (.zip):** unzip it into a folder and open VizcachaIDE. It needs no installation and no extra DLLs, only WebView2 (already on Windows 11 and almost every updated Windows 10).',
            ],
          },
          {
            p: 'Since the installer is not digitally signed yet, Windows may show *"Windows protected your PC"*. Click **More info → Run anyway**. [Here is why it is safe](/en#a-new-project).',
          },
          { h: 'macOS and Linux' },
          {
            p: 'These are previews, not yet tested on real machines. On macOS the app is not notarized: right-click it and choose **Open**. On Linux, make the AppImage executable (`chmod +x VizcachaIDE-*.AppImage`) and run it.',
          },
          { h: 'Check your download' },
          {
            p: 'Each release includes a `SHA256SUMS` file. In PowerShell, run `Get-FileHash .\\VizcachaIDE-…zip`: the value must match the one in the file.',
          },
        ],
      },
      {
        id: 'first-run',
        t: 'Your first run',
        blocks: [
          {
            p: 'When you open VizcachaIDE you will see an empty tab ready for typing. The interface uses your system language (English or Spanish).',
          },
          {
            ol: [
              'Type a program, for example the one below.',
              'Press **F5** to run it. The output appears in the bottom panel.',
              'Save the file with **Ctrl+S**.',
            ],
          },
          {
            code: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello, VizcachaIDE!")\n}',
          },
        ],
      },
      {
        id: 'run',
        t: 'Run your program',
        blocks: [
          {
            ul: [
              '**Run (F5):** builds and runs your program. **Stop (Shift+F5)** ends it whenever you want.',
              'If your program asks for keyboard input, type in the console and press Enter.',
              'You can pass arguments to your program and open projects that use `go.mod`.',
              'When you save, your code is tidied up with the standard Go format (gofmt).',
            ],
          },
        ],
      },
      {
        id: 'errors',
        t: 'Understand errors',
        blocks: [
          {
            p: 'When Go finds a problem, the **Assistant** (right-hand panel) explains what happened and how to fix it, in English or Spanish. It recognizes 25 common errors, such as a declared and unused variable, an extra import, mismatched types, an index out of range or an uninitialized map.',
          },
          {
            ul: [
              'Errors are underlined as you type, before you run.',
              "Go's original message is always shown, with a button to search for it online. That way you learn to read the real errors.",
            ],
          },
          {
            p: 'A good exercise: make a mistake on purpose (declare a variable you never use), press F5 and read the Assistant explanation.',
          },
        ],
      },
      {
        id: 'debug',
        t: 'Debug step by step',
        blocks: [
          {
            ol: [
              'Save your file (debugging needs a saved file).',
              'Click next to a line number to set a **breakpoint**.',
              'Press **Debug (F6)**. The program stops at that line.',
              'Move forward with **Next line (F7)**, **Go into function (F8)** or **Leave function (F9)**.',
              'Watch your variables at each step: the one that just changed is highlighted. The "How you got here" panel shows the call stack.',
              '**Continue (Shift+F6)** runs to the next breakpoint; **Stop debugging (Shift+F5)** ends the session.',
            ],
          },
          {
            p: 'With **Run to here (Ctrl+F10)** the program runs up to the line where the cursor is. Debugging uses Delve, which is already included in the full version.',
          },
        ],
      },
      {
        id: 'shortcuts',
        t: 'Keyboard shortcuts',
        blocks: [
          {
            table: [
              ['Run / Stop', 'F5 / Shift+F5'],
              ['Debug / Continue', 'F6 / Shift+F6'],
              ['Next line / Go into / Leave', 'F7 / F8 / F9'],
              ['Run to here', 'Ctrl+F10'],
              ['Save', 'Ctrl+S'],
              ['Find', 'Ctrl+F'],
              ['Go to line', 'Ctrl+G'],
              ['Go to definition', 'F12 or Ctrl+click'],
              ['Completion', 'Ctrl+Space'],
              ['Zoom: in / out / reset', 'Ctrl + / Ctrl − / Ctrl 0'],
            ],
            head: ['Action', 'Shortcut'],
            note: 'On macOS, use ⌘ instead of Ctrl.',
          },
        ],
      },
      {
        id: 'troubleshooting',
        t: 'Troubleshooting',
        blocks: [
          { h: 'Windows shows "Windows protected your PC"' },
          {
            p: 'This is normal for new programs without a digital signature. Click **More info → Run anyway**. You can check your download with `SHA256SUMS` and read the [source code](https://github.com/codeplai/VizcachaIDE).',
          },
          { h: 'The window does not open or appears blank (WebView2)' },
          {
            p: 'VizcachaIDE uses WebView2, which comes with Windows 11 and almost every updated Windows 10. If your computer does not have it, install the "WebView2 Runtime" from Microsoft and open VizcachaIDE again. Updating Windows usually fixes it too.',
          },
          { h: 'It says it cannot find Go' },
          {
            p: 'If you use the **lite** version, you need Go installed and available on your `PATH`. If you would rather not install it, use the **full** version, which already includes it.',
          },
          { h: 'The debugger or completion do not work' },
          { p: 'In the lite version you need to install Delve and gopls:' },
          {
            code: 'go install github.com/go-delve/delve/cmd/dlv@latest\ngo install golang.org/x/tools/gopls@latest',
          },
          { h: 'You found a bug or have an idea' },
          {
            p: 'Open an [issue on GitHub](https://github.com/codeplai/VizcachaIDE/issues) or write to [hola@codeplai.pe](mailto:hola@codeplai.pe).',
          },
        ],
      },
    ],
  },
};
