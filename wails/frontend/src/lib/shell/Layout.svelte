<script lang="ts">
  import { onMount } from 'svelte'
  import { Pane, PaneGroup, PaneResizer } from 'paneforge'
  import EditorPane from '../editor/EditorPane.svelte'
  import AssistantPanel from '../panels/AssistantPanel.svelte'
  import BottomPanel from '../panels/BottomPanel.svelte'
  import FilesPanel from '../panels/FilesPanel.svelte'
  import OutlinePanel from '../panels/OutlinePanel.svelte'
  import SearchPanel from '../panels/SearchPanel.svelte'
  import { assistantOpen, sidebarOpen, sidebarView } from '../stores'
  import '../panels/panel.css'
  import './dialog.css'
  import AboutDialog from './AboutDialog.svelte'
  import ConfirmHost from './ConfirmHost.svelte'
  import DevControls from './DevControls.svelte'
  import FirstRunWizard from './FirstRunWizard.svelte'
  import NewProjectDialog from './NewProjectDialog.svelte'
  import PackagesDialog from './PackagesDialog.svelte'
  import QuickOpen from './QuickOpen.svelte'
  import RunMemberDialog from './RunMemberDialog.svelte'
  import NoticeBar from './NoticeBar.svelte'
  import Rail from './Rail.svelte'
  import SettingsDialog from './SettingsDialog.svelte'
  import StatusBar from './StatusBar.svelte'
  import TitleBar from './TitleBar.svelte'

  // The side panel and the Assistant can be hidden (rail, title bar buttons, Ctrl+B, Ctrl+Alt+B).
  // paneforge remembers collapsed panes, so the stores start from what the panes restored.
  let sidebarPane: ReturnType<typeof Pane> | undefined = $state()
  let assistantPane: ReturnType<typeof Pane> | undefined = $state()
  let ready = $state(false)

  onMount(() => {
    if (sidebarPane) sidebarOpen.set(!sidebarPane.isCollapsed())
    if (assistantPane) assistantOpen.set(!assistantPane.isCollapsed())
    ready = true
  })

  const follow = (pane: ReturnType<typeof Pane> | undefined, open: boolean): void => {
    if (!ready || !pane) return
    if (open && pane.isCollapsed()) pane.expand()
    else if (!open && !pane.isCollapsed()) pane.collapse()
  }

  $effect(() => follow(sidebarPane, $sidebarOpen))
  $effect(() => follow(assistantPane, $assistantOpen))
</script>

<div class="window">
  <DevControls />
  <TitleBar />
  <div class="middle">
    <Rail />
    <PaneGroup direction="horizontal" autoSaveId="vizcacha-main">
      <Pane
        bind:this={sidebarPane}
        defaultSize={17}
        minSize={12}
        maxSize={35}
        collapsible
        collapsedSize={0}
        onCollapse={() => sidebarOpen.set(false)}
        onExpand={() => sidebarOpen.set(true)}
      >
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
      <Pane
        bind:this={assistantPane}
        defaultSize={23}
        minSize={16}
        maxSize={40}
        collapsible
        collapsedSize={0}
        onCollapse={() => assistantOpen.set(false)}
        onExpand={() => assistantOpen.set(true)}
      >
        <AssistantPanel />
      </Pane>
    </PaneGroup>
  </div>
  <NoticeBar />
  <StatusBar />
  <SettingsDialog />
  <AboutDialog />
  <PackagesDialog />
  <NewProjectDialog />
  <RunMemberDialog />
  <QuickOpen />
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
