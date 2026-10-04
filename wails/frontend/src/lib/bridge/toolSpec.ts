// Builds the ToolSpec of a language profile (the fields a tool does not use stay empty).
import type { ToolSpec } from '../domain'

export const tool = (spec: Partial<ToolSpec> & Pick<ToolSpec, 'id' | 'role'>): ToolSpec => ({
  labelKey: '',
  missingKey: '',
  installUrl: '',
  installCommand: '',
  providedBy: '',
  ...spec
})
