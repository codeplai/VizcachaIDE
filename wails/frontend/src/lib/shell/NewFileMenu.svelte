<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { newFile, profiles } from '../stores'
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger class="menu-item new-of-trigger">
    {$t('files.newFileOfMenu')}
    <span aria-hidden="true">▸</span>
  </DropdownMenu.SubTrigger>
  <DropdownMenu.Portal>
    <DropdownMenu.SubContent class="menu" sideOffset={4}>
      {#each $profiles as profile (profile.id)}
        <DropdownMenu.Item class="menu-item" onSelect={() => void newFile(bridge, profile.id)}>
          {$t('files.newFileOf', { values: { codeLanguage: $t(profile.nameKey) } })}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.SubContent>
  </DropdownMenu.Portal>
</DropdownMenu.Sub>

<style>
  :global(.new-of-trigger) {
    display: flex;
    justify-content: space-between;
    gap: 12px;
  }
</style>
