/** Whether two paths name the same file. Windows paths ignore case and the kind of slash. */
export const samePath = (a: string, b: string): boolean => {
  const clean = (path: string): string => path.replace(/\\/g, '/')
  const windows = /^[a-z]:|\\/i.test(a) || /^[a-z]:|\\/i.test(b)
  return windows ? clean(a).toLowerCase() === clean(b).toLowerCase() : clean(a) === clean(b)
}
