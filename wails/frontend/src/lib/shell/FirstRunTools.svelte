<script lang="ts">
  import { onMount } from 'svelte'
  import { bridge } from '../bridge'
  import type { LanguageProfile, ToolSpec } from '../domain'
  import { t } from '../i18n'
  import FirstRunCppHint from './FirstRunCppHint.svelte'
  import { enabledProfiles, refreshTools, toolStatus, tools } from '../stores'

  // The wizard's check: the runtime (or the compiler) of each language the student chose.
  const runtimeOf = (profile: LanguageProfile): ToolSpec | undefined =>
    profile.tools.find((spec) => spec.role === 'runtime' || spec.role === 'compiler')

  const checked = $derived(
    $enabledProfiles.flatMap((profile) => {
      const spec = runtimeOf(profile)
      return spec ? [{ profile, spec, version: toolStatus(spec.id, $tools)?.version ?? '' }] : []
    })
  )

  onMount(() => void refreshTools(bridge))
</script>

{#each checked as { profile, spec, version } (profile.id)}
  <h3>
    {#if profile.id === 'go'}
      {#if version}
        {$t('firstRun.step2', { values: { version } })}
      {:else}
        {$t('firstRun.checking')}
      {/if}
    {:else if version}
      {$t(profile.nameKey)} · {$t('settings.version', { values: { version } })}
    {:else}
      {$t(profile.nameKey)} · {$t('settings.notFound')}
    {/if}
  </h3>
  {#if !version}
    <p>{$t(spec.missingKey || 'errors.toolNotFound', { values: { tool: spec.id } })}</p>
    {#if profile.id === 'cpp'}<FirstRunCppHint {spec} />{/if}
  {/if}
{/each}

<style>
  h3 {
    margin: 0;
    font-size: 15px;
  }
</style>
