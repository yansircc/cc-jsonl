<script lang="ts">
  import type { SearchHit, Highlight, ContextResponse } from '$lib/types.gen';
  import { getContext } from '$lib/api';

  let { hit }: { hit: SearchHit } = $props();

  let expanded = $state(false);
  let context = $state<ContextResponse | null>(null);
  let loadingCtx = $state(false);

  function highlightSnippet(text: string, highlights: Highlight[]): string {
    if (!highlights || highlights.length === 0) return escapeHtml(text);
    const sorted = [...highlights].sort((a, b) => b.start - a.start);
    let result = text;
    for (const h of sorted) {
      const before = result.slice(0, h.start);
      const match = result.slice(h.start, h.end);
      const after = result.slice(h.end);
      result = before + `<mark>${escapeHtml(match)}</mark>` + after;
    }
    return result;
  }

  function escapeHtml(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  function formatTime(ts: string): string {
    return new Date(ts).toLocaleString('zh-CN', {
      month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit'
    });
  }

  function shortProject(path: string): string {
    const parts = path.split('/');
    return parts.slice(-2).join('/');
  }

  async function toggle() {
    if (expanded) {
      expanded = false;
      return;
    }
    expanded = true;
    if (!context) {
      loadingCtx = true;
      try {
        context = await getContext(hit.sessionId, hit.messageUuid, 1);
      } catch {
        // 静默失败
      } finally {
        loadingCtx = false;
      }
    }
  }
</script>

<div class="item">
  <a href="#expand" class="snippet" onclick={(e) => { e.preventDefault(); toggle(); }}>{@html highlightSnippet(hit.snippet, hit.highlights)}</a>
  <div class="meta">
    <span class="u">{hit.role}</span> &middot;
    {shortProject(hit.project)} &middot;
    {formatTime(hit.timestamp)}
  </div>
  {#if expanded}
    <div class="ctx">
      {#if loadingCtx}
        <p>...</p>
      {:else if context}
        {#each context.messages as msg, i}
          <div class="msg" class:hit={i === context.hitIndex}>
            <b class="u">{msg.role}:</b> {msg.text}
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>

<style lang="scss">
  .item {
    margin-bottom: 12px;
  }
  .snippet {
    text-decoration: none;
    cursor: pointer;
    white-space: pre-wrap;
    word-break: break-word;
    &:hover { text-decoration: underline; }
  }
  .meta {
    color: #888;
    font-size: 12px;
    margin-top: 2px;
  }
  .u {
    color: #060;
  }
  .ctx {
    margin: 6px 0 6px 16px;
    border-left: 2px solid #ccc;
    padding-left: 8px;
  }
  .msg {
    margin: 4px 0;
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 13px;
  }
  .msg.hit {
    background: #ff0;
  }
  @media (prefers-color-scheme: dark) {
    .meta { color: #777; }
    .u { color: #8c8; }
    .ctx { border-left-color: #444; }
    .msg.hit { background: #660; }
  }
</style>
