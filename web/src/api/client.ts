import type { StatusResponse, PostRequest, PostResponse, ParticipantsResponse } from '../types';

const API_BASE = '';

export async function fetchStatus(sessionId: string, after?: number): Promise<StatusResponse> {
  const params = new URLSearchParams({ session: sessionId });
  if (after !== undefined && after > 0) {
    params.set('after', String(after));
  }
  const response = await fetch(`${API_BASE}/api/status?${params}`);
  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || `Status fetch failed: ${response.status}`);
  }
  return response.json();
}

export async function postMessage(request: PostRequest): Promise<PostResponse> {
  const response = await fetch(`${API_BASE}/api/post`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });
  if (response.status === 409) {
    throw new Error('STALE');
  }
  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || 'Post failed');
  }
  return response.json();
}

export async function fetchParticipants(sessionId: string): Promise<ParticipantsResponse> {
  const response = await fetch(`${API_BASE}/api/participants?session=${sessionId}`);
  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.error || 'Participants fetch failed');
  }
  return response.json();
}
