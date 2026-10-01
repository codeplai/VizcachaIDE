<script lang="ts">
  import { t } from '../i18n'
</script>

<section class="side">
  <h2 class="side-h">{$t('panels.search')}</h2>
</section>

<style>
  .side {
    height: 100%;
    background: var(--chrome);
    padding: 10px 0;
  }
  .side-h {
    margin: 0;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
    padding: 2px 14px 6px;
  }
</style>
