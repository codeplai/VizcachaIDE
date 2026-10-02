# Plan: Thonny-style file handling (VizcachaIDE 2.0)

**Owner's decisions (2026-10-01):** "Delete" sends the file to the **Recycle Bin**; access is through a **File menu + 3 buttons** (New, Open, Save).

## What is missing today

| Action | Thonny | VizcachaIDE 2.0 today |
|---|---|---|
| New file (Ctrl+N) | File menu and toolbar button | Only from the first-run wizard |
| Open file (Ctrl+O) | Menu and button | **Does not exist**: only "Open folder…" under "More" |
| Open folder | Files panel | Under "More" and in the empty panel |
| Recent files | File menu | Under "More" (just added) |
| Save (Ctrl+S) | Menu and button | Shortcut only; **a new file cannot be saved** |
| Save as (Ctrl+Shift+S) | Yes | **Does not exist** |
| Save all | Yes | Does not exist |
| Close (Ctrl+W) and close all | Yes | Only through the × on each tab |
| Create, rename or delete from the Files panel | Context menu | **Does not exist** |
| Show in Explorer | Context menu | Does not exist |

The most serious gap: a program started with "blank file" can be run, but **cannot be saved**.

## Interface design

1. **A visible "File" menu** in the title bar, on the left, next to the logo. This is what a beginner expects, and Thonny has it in the same place. Contents and shortcuts:
   - New file (Ctrl+N)
   - Open file… (Ctrl+O)
   - Open folder…
   - Open recent ▸ (moves here from "More")
   - Save (Ctrl+S)
   - Save as… (Ctrl+Shift+S)
   - Save all
   - Close (Ctrl+W)
   - Close all

   "More" keeps only: Go modules, Settings and About.
2. **Three icon buttons** to the left of Run: New, Open and Save, with a tooltip and their shortcut. "Save" looks disabled when there are no changes. Run remains the primary button.
3. **Files panel**:
   - In the header, buttons for new file, new folder and refresh.
   - Right-click on a file or folder: New file here, New folder, Rename (F2), Delete (Del), Show in Explorer, Copy path.
   - When creating a file or folder, the name is typed in the tree row itself.
4. **New file**:
   - Ctrl+N opens an "untitled" tab with the minimal Go template (`package main` + `func main`), as today.
   - If a folder is open, the name is requested in the tree.
5. **Saving an untitled file**: Ctrl+S opens "Save as". Once saved, the tab becomes a real file; from then on gopls, the change watcher and the recent list work with it. Closing an untitled tab with changes offers "Save", which also opens "Save as".
6. **Delete**:
   - Asks for confirmation with the action on the button: "Delete *main.go*?" → **Delete** / **Cancel**.
   - If the file is open, its tab is closed.

## Texts (EN / ES)

They follow [UX_COPY.md](UX_COPY.md): informal "you" (tuteo in Spanish) and verbs on buttons.

- *File* / *Archivo*
- *New file* / *Nuevo archivo*
- *Open file…* / *Abrir archivo…*
- *Save as…* / *Guardar como…*
- *Save all* / *Guardar todo*
- *Close* / *Cerrar*
- *Close all* / *Cerrar todo*
- *Rename* / *Renombrar*
- *Delete* / *Eliminar*
- *Show in Explorer* / *Mostrar en el Explorador* (on macOS, *Finder*)
- *Copy path* / *Copiar ruta*
- *New folder* / *Nueva carpeta*

Errors use the structure what happened + how to fix it. For example, "A file named *main.go* already exists in this folder. Choose another name." (ES: «Ya existe un archivo llamado *main.go* en esta carpeta. Elige otro nombre.»)

## Backend (Go)

`FilesService` gains these methods. They follow the same architecture rules: native dialogs through the Wails runtime, only in `bridge`.

- `OpenFileDialog() (string, error)`: native dialog with a "Go files (*.go)" and "All files" filter.
- `SaveFileDialog(suggestedName, folder string) (string, error)`: native dialog; adds `.go` if missing.
- `CreateFile(path, text string) error` and `CreateFolder(path string) error`: fail if the path already exists.
- `Rename(from, to string) error`: fails if the destination exists.
- `Delete(path string) error`: sends to the Recycle Bin. On Windows it uses `SHFileOperationW` with `FOF_ALLOWUNDO` (shell32, via golang.org/x/sys/windows). On macOS it uses Finder through `osascript`. On Linux it uses `gio trash`. If the Recycle Bin is not available, it returns a clear error and does not delete.
- `RevealInExplorer(path string) error`: opens the file manager with the file selected (`explorer /select,` on Windows, `open -R` on macOS, `xdg-open` on the folder on Linux).

The operations live in `app` (use cases with name validation) and the file system in an adapter. They come with tests in temporary folders.

## Frontend

- `stores/fileCommands.ts`: new, open, save as, save all, close all.
- `stores/fileTreeCommands.ts`: create, rename and delete in the tree. Renaming or deleting updates the open tabs, buffers, gopls, the watcher and the recent list.
- Components:
  - `shell/FileMenu.svelte`: bits-ui DropdownMenu.
  - `shell/FileButtons.svelte`.
  - `panels/FileTreeMenu.svelte`: bits-ui ContextMenu.
  - Inline name editing inside `FileTreeNode.svelte`.
- Shortcuts in `shell/shortcuts.ts`: Ctrl+N, Ctrl+O, Ctrl+Shift+S, Ctrl+W and F2/Del in the tree. All with `preventDefault`, so WebView2 does not open windows or close anything.
- Bridge mock for `npm run dev` and tests with vitest.

## Execution

Two agents in parallel, in separate copies of the repo, after the agents still working (console and quick improvements) finish:

| Agent | Scope |
|---|---|
| A: File menu and save | "File" menu, buttons, New, Open file, Save as (including untitled), Save all, Close and Close all, shortcuts, backend dialog methods |
| B: Files panel | Header buttons, context menu, create, rename, delete, show in Explorer, copy path, inline editing, backend file methods |

So that they do not collide, I first land the contract on `main`: the new `FilesService` signatures in Go, `types.ts` and the mock.

## Verification

- Go tests and vitest, lint and the architecture test.
- Test in the real app (`wails dev`):
  - create a new file, run it and save it as `hola.go`;
  - open a loose file;
  - rename and delete from the tree;
  - Ctrl+W and Close all with unsaved changes;
  - "Show in Explorer".
- Screenshots in Spanish and English, and at 1024 px width.
- Update the manual (artifact) and `docs/wails/QA_WAILS.md`.
