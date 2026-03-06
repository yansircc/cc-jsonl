import type {
	SearchResult,
	ContextResponse,
	StatsResponse
} from './types.gen';

const BASE = '/api';

async function get<T>(path: string, params?: Record<string, string>): Promise<T> {
	const url = new URL(path, window.location.origin);
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			if (v) url.searchParams.set(k, v);
		}
	}
	const res = await fetch(url);
	if (!res.ok) throw new Error(`API error: ${res.status}`);
	return res.json();
}

export function searchMessages(
	q: string,
	project?: string,
	offset = 0,
	limit = 20
): Promise<SearchResult> {
	return get<SearchResult>(`${BASE}/search`, {
		q,
		project: project || '',
		offset: String(offset),
		limit: String(limit)
	});
}

export function getContext(
	sessionId: string,
	uuid: string,
	around = 1
): Promise<ContextResponse> {
	return get<ContextResponse>(`${BASE}/context/${sessionId}`, {
		uuid,
		around: String(around)
	});
}

export function getStats(): Promise<StatsResponse> {
	return get<StatsResponse>(`${BASE}/stats`);
}
