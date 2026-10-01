<script lang="ts">
  import { Dialog } from 'bits-ui'
  import { bridge } from '../bridge'
  import type { LanguageSetting } from '../domain'
  import { t } from '../i18n'
  import { FIRST_RUN_STEPS, completeFirstRun, settings, toolchain, updateSettings } from '../stores'

  let step = $state(1)

  const languages: { id: LanguageSetting; label: string }[] = [
    { id: 'en', label: 'language.en' },
    { id: 'es', label: 'language.es' }
  ]
  const goVersion = $derived($toolchain?.goVersion ?? '')
</script>

<Dialog.Root open={$settings?.firstRun === true}>
  <Dialog.Portal>
    <Dialog.Overlay class="dlg-overlay" />
    <Dialog.Content class="dlg" interactOutsideBehavior="ignore" escapeKeydownBehavior="ignore">
      <Dialog.Title>{$t('firstRun.welcome')}</Dialog.Title>
      <Dialog.Description class="step">
        {$t('firstRun.step', { values: { current: step, total: FIRST_RUN_STEPS } })}
      </Dialog.Description>

      {#if step === 1}
        <h3>{$t('firstRun.step1')}</h3>
        <div class="choices" role="group" aria-label={$t('firstRun.step1')}>
          {#each languages as option (option.id)}
            <button
              type="button"
              class="dlg-button"
              class:primary={$settings?.language === option.id}
              aria-pressed={$settings?.language === option.id}
              onclick={() => updateSettings(bridge, { language: option.id })}
            >
              {$t(option.label)}
            </button>
          {/each}
        </div>
      {:else if step === 2}
        <h3>
          {#if goVersion}
            {$t('firstRun.step2', { values: { version: goVersion } })}
          {:else}
            {$t('firstRun.checking')}
          {/if}
        </h3>
        {#if !goVersion}<p>{$t('errors.goNotFound')}</p>{/if}
      {:else}
        <h3>{$t('firstRun.step3')}</h3>
        <div class="choices">
          <button
            type="button"
            class="dlg-button primary"
            onclick={() => completeFirstRun(bridge, 'hello')}
          >
            {$t('firstRun.openHello')}
          </button>
          <button
            type="button"
            class="dlg-button"
            onclick={() => completeFirstRun(bridge, 'blank')}
          >
            {$t('firstRun.startBlank')}
          </button>
        </div>
      {/if}

      <div class="dlg-actions">
        {#if step > 1}
          <button type="button" class="dlg-button" onclick={() => step--}>
            {$t('firstRun.back')}
          </button>
        {/if}
        {#if step < FIRST_RUN_STEPS}
          <button type="button" class="dlg-button primary" onclick={() => step++}>
            {$t('firstRun.next')}
          </button>
        {/if}
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<style>
  h3 {
    margin: 0;
    font-size: 15px;
  }
  .choices {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
</style>
