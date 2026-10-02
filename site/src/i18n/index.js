import es from './es.js';
import en from './en.js';

export const languages = { es, en };
export const defaultLang = 'es';

/** Datos del proyecto: editar aquí y cambia en todo el sitio. */
export const project = {
  name: 'VizcachaIDE',
  version: '2.0.0-rc1',
  email: 'hola@codeplai.pe',
  parentUrl: 'https://codeplai.pe',
  repoUrl: 'https://github.com/codeplai/VizcachaIDE',
  releasesUrl: 'https://github.com/codeplai/VizcachaIDE/releases',
  issuesUrl: 'https://github.com/codeplai/VizcachaIDE/issues',
  year: 2026,
  legalName: 'CODEPLAI GAMES SAC',
};

/** Rutas equivalentes en cada idioma (misma convención que codeplai.pe: ES en la raíz, EN bajo /en). */
export const routes = {
  es: { home: '/', manual: '/manual' },
  en: { home: '/en', manual: '/en/manual' },
};

/** Devuelve el diccionario de copy para un idioma. */
export function t(lang) {
  return languages[lang] ?? languages[defaultLang];
}
