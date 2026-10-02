// @ts-check
import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

// El dominio se usa para el sitemap, los hreflang y las etiquetas Open Graph.
const SITE = 'https://vizcacha.codeplai.pe';

export default defineConfig({
  site: SITE,
  trailingSlash: 'never',
  build: {
    format: 'file',
  },
  integrations: [
    sitemap({
      i18n: {
        defaultLocale: 'es',
        locales: { es: 'es-PE', en: 'en' },
      },
    }),
  ],
});
