<script lang="ts">
  import type { SearchResult } from '$lib/types.gen';
  import ResultItem from './ResultItem.svelte';

  let { result, onLoadMore }: {
    result: SearchResult;
    onLoadMore?: () => void;
  } = $props();
</script>

<div>
  <p style="color: #888; font-size: 12px; margin-bottom: 8px">
    {result.total} results, {result.queryTimeMs}ms
  </p>
  {#each result.results as hit, i (i)}
    <ResultItem {hit} />
  {/each}
  {#if result.hasMore && onLoadMore}
    <a href="#more" onclick={(e) => { e.preventDefault(); onLoadMore?.(); }}>
      [{result.results.length + 1}-{Math.min(result.results.length + 20, result.total)}]
    </a>
  {/if}
</div>
