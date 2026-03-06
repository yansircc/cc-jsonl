<script lang="ts">
  import ResultList from '$lib/components/ResultList.svelte';
  import { searchMessages, getStats } from '$lib/api';
  import type { SearchResult, StatsResponse } from '$lib/types.gen';

  let query = $state('');
  let input = $state('');
  let result = $state<SearchResult | null>(null);
  let loading = $state(false);
  let error = $state('');
  let stats = $state<StatsResponse | null>(null);

  // 加载统计信息
  $effect(() => {
    getStats().then(s => stats = s).catch(() => {});
  });

  async function handleSearch() {
    const q = input.trim();
    if (!q) return;
    query = q;
    loading = true;
    error = '';
    try {
      result = await searchMessages(q);
    } catch (e) {
      error = e instanceof Error ? e.message : '搜索失败';
      result = null;
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    if (!result || !result.hasMore) return;
    loading = true;
    try {
      const more = await searchMessages(query, undefined, result.results.length);
      result = {
        ...more,
        results: [...result.results, ...more.results]
      };
    } catch (e) {
      error = e instanceof Error ? e.message : '加载失败';
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === '/' && !(e.target instanceof HTMLInputElement)) {
      e.preventDefault();
      document.querySelector<HTMLInputElement>('input[name="q"]')?.focus();
    }
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB';
    return (bytes / 1024 / 1024 / 1024).toFixed(1) + ' GB';
  }

  const hasResult = $derived(result !== null);
</script>

<svelte:window onkeydown={handleKeydown} />

{#if !hasResult}
  <div class="landing">
    <b>cc-jsonl</b>
    <form onsubmit={(e) => { e.preventDefault(); handleSearch(); }}>
      <input name="q" type="text" bind:value={input} placeholder="搜索对话记录..." autofocus>
    </form>
    {#if stats}
      <span class="hint">{stats.sessions} sessions, {formatSize(stats.totalSize)}</span>
    {/if}
  </div>
{:else}
  <div class="top">
    <b><a href="/" onclick={(e) => { e.preventDefault(); result = null; input = ''; query = ''; }}>cc-jsonl</a></b>
    <form onsubmit={(e) => { e.preventDefault(); handleSearch(); }}>
      <input name="q" type="text" bind:value={input} autofocus>
    </form>
  </div>

  {#if error}
    <p style="color: #c00">{error}</p>
  {/if}

  <ResultList {result} onLoadMore={result.hasMore ? loadMore : undefined} />
{/if}

{#if loading}
  <p style="color: #888">搜索中...</p>
{/if}

<style lang="scss">
  .landing {
    text-align: center;
    margin-top: 30vh;
    b { font-size: 18px; }
    form {
      margin: 12px 0 8px;
      input {
        width: 400px;
        max-width: 80vw;
        padding: 4px;
        border: 1px solid #888;
      }
    }
    .hint { color: #888; font-size: 12px; }
  }
  .top {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
    b a {
      color: inherit;
      text-decoration: none;
    }
    form {
      flex: 1;
      input {
        width: 100%;
        padding: 4px;
        border: 1px solid #888;
      }
    }
  }
</style>
