# Sitio web de VizcachaIDE (vizcacha.codeplai.pe)

Sitio estático en **Astro 7**, bilingüe (ES en la raíz, EN bajo `/en`), pensado para
**Cloudflare Pages**. Usa la misma base que codeplai.pe (Geist/Inter, tokens de color,
patrón de i18n, cabecera y pie), con el cian de Go (`#00ADD8`) como identidad propia.

## Estructura

```
site/
├─ astro.config.mjs        sitio, sitemap con i18n
├─ public/                 favicon, og.png (1200x630), robots.txt, _headers, llms.txt
└─ src/
   ├─ i18n/                es.js, en.js (todo el copy), index.js (rutas y datos del proyecto)
   ├─ layouts/Base.astro   <head>, SEO, hreflang, OG/Twitter, tema claro/oscuro
   ├─ components/          Header, Footer, Hero, Features, Download, SafetyNotice, ManualBody…
   ├─ assets/              capturas (ES/EN), logo y mascota (se optimizan con astro:assets)
   ├─ pages/               index, manual, en/index, en/manual
   └─ styles/global.css
```

Para cambiar textos, edita `src/i18n/es.js` y `src/i18n/en.js`. Versión, enlaces y correo
están en `src/i18n/index.js`. Las capturas son copias de `docs/images/wails/`: si cambian,
vuelve a copiarlas a `src/assets/`.

Rutas: `/`, `/manual`, `/en`, `/en/manual` (sin barra final, igual que codeplai.pe).

## Desarrollo local

```bash
cd site
npm install
npm run dev        # http://localhost:4321
npm run build      # genera dist/
npm run preview    # sirve dist/
```

Requiere Node 22 (`.nvmrc`).

## Publicar en Cloudflare Pages

1. En **dash.cloudflare.com → Compute (Workers y Pages) → Crear → Pages → Conectar a Git**,
   autoriza GitHub y elige el repositorio `codeplai/VizcachaIDE`.
2. Configuración de compilación:

   | Campo | Valor |
   |---|---|
   | Framework preset | Astro |
   | Directorio raíz (Root directory) | `site` |
   | Comando de compilación | `npm run build` |
   | Directorio de salida | `dist` |
   | Versión de Node | `22` (lo toma de `site/.nvmrc`; si hace falta, variable `NODE_VERSION=22`) |

3. **Guardar e implementar.** En unos minutos tendrás una URL tipo `vizcachaide.pages.dev`.
   Cada `git push` a la rama de producción vuelve a publicar el sitio.
4. En el proyecto de Pages → **Custom domains → Set up a domain**, añade
   `vizcacha.codeplai.pe`. Como la zona DNS de `codeplai.pe` está en Cloudflare, el registro
   CNAME y el certificado HTTPS se crean solos (suele tardar unos minutos).
5. Comprueba `https://vizcacha.codeplai.pe`, `/en`, `/manual`, `/sitemap-index.xml` y
   `/robots.txt`.

Opcional: en **Settings → Builds → Build watch paths** limita las compilaciones a `site/*`
para que los cambios del IDE no republiquen el sitio.

## SEO incluido

Títulos y descripciones por idioma, `canonical`, `hreflang` (es-PE, en, x-default), Open Graph
y Twitter con `og.png`, datos estructurados `SoftwareApplication`, `sitemap-index.xml`,
`robots.txt` (misma política que codeplai.pe) y `llms.txt`.
