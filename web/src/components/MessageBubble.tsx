import Markdown from 'react-markdown';
import type { APIEvent } from '../types';
import { formatTimestamp } from '../utils/time';

interface MessageBubbleProps {
  event: APIEvent;
}

export function MessageBubble({ event }: MessageBubbleProps) {
  const isModerator = event.participant === 'Moderator';

  return (
    <div
      className={`mb-4 rounded-lg p-4 ${
        isModerator
          ? 'border-2 border-blue-500 bg-blue-50 dark:border-blue-400 dark:bg-blue-950'
          : 'border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800'
      }`}
    >
      <div className="mb-2 flex items-center justify-between">
        <span className={`font-semibold ${isModerator ? 'text-blue-700 dark:text-blue-300' : 'text-gray-900 dark:text-gray-100'}`}>
          {event.participant}
        </span>
        <div className="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500">
          <span>{formatTimestamp(event.timestamp_millis)}</span>
          <span>#{event.number}</span>
        </div>
      </div>
      <div className="markdown-content text-gray-800 dark:text-gray-200">
        <Markdown>{event.content || ''}</Markdown>
      </div>
      {event.next && (
        <div className="mt-2 text-right text-sm text-gray-500 dark:text-gray-400">
          → {event.next}
        </div>
      )}
    </div>
  );
}
