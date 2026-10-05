// A few fake packages per language so `npm run dev` shows the search list without a network.
// Typing "down" in the demo behaves like an index that cannot be reached.
import type { CodeLanguage, PackageInfo } from '../domain'

const info = (name: string, version: string, description: string, url: string): PackageInfo => ({
  name,
  version,
  description,
  url
})

const DEMO: Partial<Record<CodeLanguage, PackageInfo[]>> = {
  python: [
    info(
      'numpy',
      '2.1.0',
      'Fundamental package for array computing in Python',
      'https://pypi.org/project/numpy/'
    ),
    info(
      'numpy-financial',
      '1.0.0',
      'Simple financial functions for NumPy',
      'https://pypi.org/project/numpy-financial/'
    ),
    info('requests', '2.32.3', 'Python HTTP for Humans.', 'https://pypi.org/project/requests/'),
    info('uuid6', '2024.7.10', 'New time-based UUID formats', 'https://pypi.org/project/uuid6/')
  ],
  go: [
    info(
      'github.com/google/uuid',
      'v1.6.0',
      'Package uuid generates and inspects UUIDs.',
      'https://pkg.go.dev/github.com/google/uuid'
    ),
    info(
      'github.com/gofrs/uuid/v5',
      'v5.3.0',
      'Package uuid provides implementations of the UUID.',
      'https://pkg.go.dev/github.com/gofrs/uuid/v5'
    ),
    info(
      'math/rand',
      '',
      'Package rand implements pseudo-random number generators.',
      'https://pkg.go.dev/math/rand'
    )
  ],
  rust: [
    info(
      'rand',
      '0.10.3',
      'Random number generators and other randomness functionality.',
      'https://crates.io/crates/rand'
    ),
    info(
      'serde',
      '1.0.219',
      'A generic serialization/deserialization framework',
      'https://crates.io/crates/serde'
    ),
    info(
      'uuid',
      '1.11.0',
      'A library to generate and parse UUIDs.',
      'https://crates.io/crates/uuid'
    )
  ]
}

/** The demo packages whose name or description contains the query; rejects for "down". */
export const searchDemoIndex = async (
  codeLanguage: CodeLanguage,
  query: string
): Promise<PackageInfo[]> => {
  const wanted = query.trim().toLowerCase()
  if (wanted === 'down') throw new Error('the package index could not be searched')
  if (wanted === '') return []
  return (DEMO[codeLanguage] ?? []).filter(
    (found) =>
      found.name.toLowerCase().includes(wanted) || found.description.toLowerCase().includes(wanted)
  )
}
