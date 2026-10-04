<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { chooseMember, dismissMemberChoice, memberChoice } from '../stores'
  import Modal from './Modal.svelte'
</script>

<Modal open={$memberChoice !== null} title={$t('run.chooseMember')} onClose={dismissMemberChoice}>
  <ul class="members">
    {#each $memberChoice?.members ?? [] as member (member)}
      <li>
        <button
          type="button"
          class="dlg-button"
          data-member={member}
          onclick={() => chooseMember(bridge, member)}
        >
          {member}
        </button>
      </li>
    {/each}
  </ul>
</Modal>

<style>
  .members {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
