<script lang="ts">
  import { bridge } from '../bridge'
  import type { CodeLanguage } from '../domain'
  import { t } from '../i18n'
  import {
    canCreateProject,
    chooseProjectLocation,
    closeNewProject,
    createNewProject,
    enabledProfiles,
    newProjectBusy,
    newProjectError,
    newProjectLanguage,
    newProjectLocation,
    newProjectName,
    openDialog,
    visibleNameProblem
  } from '../stores'
  import Modal from './Modal.svelte'

  const problem = $derived($visibleNameProblem ?? $newProjectError?.key ?? null)
  const problemText = $derived(
    problem ? $t(problem, { values: { reason: $newProjectError?.reason ?? '' } }) : ''
  )
</script>

<Modal open={$openDialog === 'newProject'} title={$t('project.title')} onClose={closeNewProject}>
  <form
    id="new-project-form"
    class="project-form"
    onsubmit={(event) => {
      event.preventDefault()
      void createNewProject(bridge)
    }}
  >
    <div class="field">
      <label for="project-name">{$t('project.name')}</label>
      <input
        id="project-name"
        type="text"
        autocomplete="off"
        bind:value={$newProjectName}
        oninput={() => newProjectError.set(null)}
        aria-invalid={$visibleNameProblem !== null}
      />
    </div>
    <div class="field">
      <label for="project-language">{$t('project.language')}</label>
      <select id="project-language" bind:value={$newProjectLanguage}>
        {#each $enabledProfiles as profile (profile.id)}
          <option value={profile.id as CodeLanguage}>{$t(profile.nameKey)}</option>
        {/each}
      </select>
      <span class="hint">{$t(`project.files.${$newProjectLanguage}`)}</span>
    </div>
    <div class="field">
      <label for="project-location">{$t('project.location')}</label>
      <div class="row">
        <input
          id="project-location"
          type="text"
          readonly
          value={$newProjectLocation}
          placeholder={$t('project.locationNone')}
        />
        <button type="button" class="dlg-button" onclick={() => chooseProjectLocation(bridge)}>
          {$t('project.choose')}
        </button>
      </div>
    </div>
    {#if problem}
      <p class="problem" role="alert">{problemText}</p>
    {/if}
  </form>
  {#snippet actions()}
    <button type="button" class="dlg-button" onclick={closeNewProject}>
      {$t('project.cancel')}
    </button>
    <button
      type="submit"
      form="new-project-form"
      class="dlg-button primary"
      disabled={!$canCreateProject || $newProjectBusy}
    >
      {$t('project.create')}
    </button>
  {/snippet}
</Modal>

<style>
  .project-form {
    display: grid;
    gap: 14px;
  }
  .row {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .row input {
    flex: 1;
  }
  .problem {
    color: var(--err);
  }
</style>
