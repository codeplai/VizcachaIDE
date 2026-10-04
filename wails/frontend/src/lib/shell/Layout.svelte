<script lang="ts">
  import { Pane, PaneGroup, PaneResizer } from 'paneforge'
  import EditorPane from '../editor/EditorPane.svelte'
  import AssistantPanel from '../panels/AssistantPanel.svelte'
  import BottomPanel from '../panels/BottomPanel.svelte'
  import FilesPanel from '../panels/FilesPanel.svelte'
  import OutlinePanel from '../panels/OutlinePanel.svelte'
  import SearchPanel from '../panels/SearchPanel.svelte'
  import { sidebarView } from '../stores'
  import '../panels/panel.css'
  import './dialog.css'
  import AboutDialog from './AboutDialog.svelte'
  import ConfirmHost from './ConfirmHost.svelte'
  import DevControls from './DevControls.svelte'
  import FirstRunWizard from './FirstRunWizard.svelte'
  import PackagesDialog from './PackagesDialog.svelte'
  import NoticeBar from './NoticeBar.svelte'
  import Rail from './Rail.svelte'
  import SettingsDialog from './SettingsDialog.svelte'
  import StatusBar from './StatusBar.svelte'
  import TitleBar from './TitleBar.svelte'
</script>

<div class="window">
  <DevControls />
  <TitleBar />
  <div class="middle">
    <Rail />
    <PaneGroup direction="horizontal" autoSaveId="vizcacha-main">
      <Pane defaultSize={17} minSize={12} maxSize={35}>
        {#if $sidebarView === 'files'}
          <FilesPanel />
        {:else if $sidebarView === 'outline'}
          <OutlinePanel />
        {:else}
          <SearchPanel />
        {/if}
      </Pane>
      <PaneResizer class="resizer" />
      <Pane defaultSize={60} minSize={30}>
        <PaneGroup direction="vertical" autoSaveId="vizcacha-center">
          <Pane defaultSize={74} minSize={30}><EditorPane /></Pane>
          <PaneResizer class="resizer horizontal" />
          <Pane defaultSize={26} minSize={12}><BottomPanel /></Pane>
        </PaneGroup>
      </Pane>
      <PaneResizer class="resizer" />
      <Pane defaultSize={23} minSize={16} maxSize={40}><AssistantPanel /></Pane>
    </PaneGroup>
  </div>
  <NoticeBar />
  <StatusBar />
  <SettingsDialog />
  <AboutDialog />
  <PackagesDialog />
  <ConfirmHost />
  <FirstRunWizard />
</div>

<style>
  .window {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--win);
  }
  .middle {
    flex: 1;
    min-height: 0;
    display: flex;
  }
  .middle :global([data-pane-group]) {
    flex: 1;
    min-width: 0;
  }
  .middle :global(.resizer) {
    width: 1px;
    background: var(--line);
    position: relative;
    flex: none;
  }
  .middle :global(.resizer.horizontal) {
    width: auto;
    height: 1px;
  }
  .middle :global(.resizer)::after {
    content: '';
    position: absolute;
    inset: 0 -3px;
  }
  .middle :global(.resizer.horizontal)::after {
    inset: -3px 0;
  }
  .middle :global(.resizer[data-active]),
  .middle :global(.resizer:hover) {
    background: var(--go);
  }
</style>
