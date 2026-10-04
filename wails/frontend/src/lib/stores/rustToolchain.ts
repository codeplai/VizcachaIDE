// What the IDE tells a student about their Rust install: how to get it and what to fix.
import { writable } from 'svelte/store'
import type { SystemPlatform } from '../platform'

export const RUSTUP_URL = 'https://rustup.rs'

const WINDOWS_INSTALL =
  'rustup-init.exe --default-host x86_64-pc-windows-gnu --default-toolchain stable -y'
const UNIX_INSTALL = "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"

/** The command that installs Rust: the GNU toolchain on Windows (the MSVC one needs a paid linker). */
export const rustupInstallCommand = (platform: SystemPlatform): string =>
  platform === 'windows' ? WINDOWS_INSTALL : UNIX_INSTALL

/** The non-blocking advice keys the backend reports, each with the command that fixes it. */
export const RUST_ADVICE_COMMANDS: Record<string, string> = {
  'errors.rustMsvcHost': 'rustup default stable-x86_64-pc-windows-gnu',
  'errors.rustNotStable': 'rustup update stable',
  'errors.rustTooOld': 'rustup update stable',
  'errors.rustupNoToolchain': 'rustup default stable'
}

/**
 * The advice keys about the Rust toolchain (errors.rustMsvcHost, errors.rustNotStable...). Empty
 * until the backend fills it (CCR: ToolStatus.Advice or a dedicated call); keys it does not
 * know are ignored when shown.
 */
export const rustAdvice = writable<string[]>([])
