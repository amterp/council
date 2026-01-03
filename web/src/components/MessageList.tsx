import { useEffect, useRef } from 'react';
import type { APIEvent } from '../types';
import { MessageBubble } from './MessageBubble';
import { EventNotice } from './EventNotice';

interface MessageListProps {
  events: APIEvent[];
}

export function MessageList({ events }: MessageListProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const prevEventCountRef = useRef(0);

  // Auto-scroll to bottom when new events arrive
  useEffect(() => {
    if (events.length > prevEventCountRef.current && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
    prevEventCountRef.current = events.length;
  }, [events.length]);

  return (
    <div ref={containerRef} className="flex-1 overflow-y-auto bg-gray-50 p-4 dark:bg-gray-950">
      {events.length === 0 ? (
        <div className="flex h-full items-center justify-center text-gray-500 dark:text-gray-400">
          No events yet
        </div>
      ) : (
        events.map((event) =>
          event.type === 'message' ? (
            <MessageBubble key={event.number} event={event} />
          ) : (
            <EventNotice key={event.number} event={event} />
          )
        )
      )}
    </div>
  );
}
