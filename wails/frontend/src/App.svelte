<script lang="ts">
  import { onMount } from 'svelte'
  import { bridge, mockControls } from './lib/bridge'
  import { parseDevQuery } from './lib/devQuery'
  import { applyLanguage, locale } from './lib/i18n'
  import Layout from './lib/shell/Layout.svelte'
  import { registerShortcuts } from './lib/shell/shortcuts'
  import { settings } from './lib/stores'
  import { startApp } from './lib/stores/startup'
  import { applyTheme } from './lib/theme/theme'

  let ready = $state(false)

  // Settings drive the language and the theme.
  $effect(() => {
    if ($settings) {
      applyLanguage($settings.language)
      applyTheme($settings.theme)
    }
  })

  onMount(() => {
    const stopShortcuts = registerShortcuts(bridge)
    let stopApp: (() => void) | undefined
    void startApp(bridge, mockControls, parseDevQuery(location.search)).then((stop) => {
      stopApp = stop
      ready = true
    })
    return () => {
      stopShortcuts()
      stopApp?.()
    }
  })
</script>

{#if ready && $locale}
  <Layout />
{/if}
