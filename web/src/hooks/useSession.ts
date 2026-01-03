import { useState, useEffect, useCallback, useRef } from 'react';
import type { APIEvent } from '../types';
import { fetchStatus } from '../api/client';

const POLL_INTERVAL = 1000;

interface UseSessionResult {
  events: APIEvent[];
  participants: string[];
  sessionId: string;
  eventCount: number;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useSession(sessionId: string): UseSessionResult {
  const [events, setEvents] = useState<APIEvent[]>([]);
  const [participants, setParticipants] = useState<string[]>([]);
  const [eventCount, setEventCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const lastEventNumRef = useRef(0);
  const initialFetchDoneRef = useRef(false);

  const poll = useCallback(async () => {
    try {
      const data = await fetchStatus(sessionId, lastEventNumRef.current);

      if (data.events.length > 0) {
        setEvents((prev) => [...prev, ...data.events]);
      }

      setParticipants(data.participants);
      setEventCount(data.event_count);
      lastEventNumRef.current = data.event_count;
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    }
  }, [sessionId]);

  const refetch = useCallback(async () => {
    await poll();
  }, [poll]);

  useEffect(() => {
    // Reset state when sessionId changes
    setEvents([]);
    setParticipants([]);
    setEventCount(0);
    setLoading(true);
    setError(null);
    lastEventNumRef.current = 0;
    initialFetchDoneRef.current = false;

    const initialFetch = async () => {
      try {
        const data = await fetchStatus(sessionId);
        setEvents(data.events);
        setParticipants(data.participants);
        setEventCount(data.event_count);
        lastEventNumRef.current = data.event_count;
        setLoading(false);
        initialFetchDoneRef.current = true;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
        setLoading(false);
      }
    };

    initialFetch();

    const interval = setInterval(() => {
      if (initialFetchDoneRef.current) {
        poll();
      }
    }, POLL_INTERVAL);

    return () => clearInterval(interval);
  }, [sessionId, poll]);

  return { events, participants, sessionId, eventCount, loading, error, refetch };
}
