<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { relativePath } from '../search/skip'
  import {
    folderRoot,
    goToLocation,
    replaceInFile,
    replaceMatch,
    searchState,
    baseName
  } from '../stores'
  import type { SearchMatch } from '../domainSearch'
  import { characterColumn } from '../search/matcher'

  /** The line cut around the match: leading blanks go, the match is marked. */
  const parts = (match: SearchMatch): { before: string; hit: string; after: string } => {
    const start = match.column - 1
    const lead = match.text.length - match.text.trimStart().length
    const cut = Math.max(0, Math.min(lead, start))
    return {
      before: match.text.slice(cut, start),
      hit: match.text.slice(start, start + match.length),
      after: match.text.slice(start + match.length)
    }
  }
</script>

<div class="results" role="region" aria-label={$t('search.results')}>
  {#each $searchState.files as file (file.path)}
    <div class="file">
      <div class="file-head">
        <span class="file-name">{baseName(file.path)}</span>
        <span class="file-dir">{relativePath(folderRoot(), file.path)}</span>
        <span class="count">{file.matches.length}</span>
        {#if $searchState.showReplace}
          <button
            type="button"
            class="mini"
            title={$t('search.replaceFile')}
            aria-label={$t('search.replaceFileLabel', { values: { file: baseName(file.path) } })}
            onclick={() => void replaceInFile(bridge, file)}>⇄</button
          >
        {/if}
      </div>
      <ul class="matches">
        {#each file.matches as match (`${match.line}:${match.column}`)}
          {@const view = parts(match)}
          <li class="match">
            <button
              type="button"
              class="go"
              onclick={() =>
                void goToLocation(bridge, {
                  file: file.path,
                  line: match.line,
                  column: characterColumn(match)
                })}
            >
              <span class="line">{match.line}</span>
              <span class="text">{view.before}<mark>{view.hit}</mark>{view.after}</span>
            </button>
            {#if $searchState.showReplace}
              <button
                type="button"
                class="mini"
                title={$t('search.replaceOne')}
                aria-label={$t('search.replaceOneLabel', {
                  values: { file: baseName(file.path), line: match.line }
                })}
                onclick={() => void replaceMatch(bridge, file.path, match)}>⇄</button
              >
            {/if}
          </li>
        {/each}
      </ul>
    </div>
  {/each}
</div>

<style>
  .results {
    overflow: auto;
    min-height: 0;
    flex: 1;
    padding: 0 6px 8px;
  }
  .file-head {
    display: flex;
    gap: 6px;
    align-items: baseline;
    padding: 6px 6px 2px;
    font: 600 13px var(--ui);
  }
  .file-dir {
    color: var(--muted);
    font-weight: 400;
    font-size: 11.5px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }
  .count {
    color: var(--muted);
    font-size: 11.5px;
  }
  .matches {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .match {
    display: flex;
    align-items: center;
  }
  .go {
    flex: 1;
    min-width: 0;
    display: flex;
    gap: 8px;
    text-align: left;
    border: 0;
    background: none;
    color: var(--ink);
    padding: 2px 6px;
    border-radius: 4px;
    cursor: pointer;
    font: 12.5px var(--mono, monospace);
  }
  .go:hover,
  .go:focus-visible {
    background: var(--go-soft, rgba(0, 120, 200, 0.12));
  }
  .line {
    color: var(--muted);
    min-width: 2.2em;
    text-align: right;
  }
  .text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: pre;
  }
  mark {
    background: rgba(255, 190, 0, 0.35);
    color: inherit;
    border-radius: 2px;
  }
  .mini {
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    border-radius: 5px;
    padding: 0 6px;
    font: 12px var(--ui);
    cursor: pointer;
    flex: none;
  }
  .mini:hover,
  .mini:focus-visible {
    border-color: var(--go);
  }
</style>
