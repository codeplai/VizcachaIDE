<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { enabledProfiles, profiles, toggleEnabledLanguage } from '../stores'

  /** Accessible name of the group (the visible heading is the caller's). */
  let { label }: { label: string } = $props()

  const isOn = (id: string): boolean => $enabledProfiles.some((profile) => profile.id === id)
  // One language must stay: its box cannot be cleared.
  const isLast = (id: string): boolean => $enabledProfiles.length === 1 && isOn(id)
</script>

<div class="choices" role="group" aria-label={label}>
  {#each $profiles as profile (profile.id)}
    <label>
      <input
        type="checkbox"
        checked={isOn(profile.id)}
        disabled={isLast(profile.id)}
        onchange={() => toggleEnabledLanguage(bridge, profile.id)}
      />
      {$t(profile.nameKey)}
    </label>
  {/each}
</div>

<style>
  .choices {
    display: grid;
    gap: 8px;
  }
  label {
    display: flex;
    gap: 8px;
    align-items: center;
    font-size: 14px;
  }
</style>
