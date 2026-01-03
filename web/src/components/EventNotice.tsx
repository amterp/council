import type { APIEvent } from '../types';
import { formatTimestamp } from '../utils/time';

interface EventNoticeProps {
  event: APIEvent;
}

export function EventNotice({ event }: EventNoticeProps) {
  let text = '';
  let icon = '';

  switch (event.type) {
    case 'session_created':
      text = `Session created`;
      icon = '🎉';
      break;
    case 'joined':
      text = `${event.participant} joined`;
      icon = '→';
      break;
    case 'left':
      text = `${event.participant} left`;
      icon = '←';
      break;
    default:
      return null;
  }

  return (
    <div className="flex items-center justify-center py-2">
      <div className="flex items-center gap-2 rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-600 dark:bg-gray-800 dark:text-gray-400">
        <span className="text-gray-400 dark:text-gray-500">{formatTimestamp(event.timestamp_millis)}</span>
        <span className="text-gray-400 dark:text-gray-500">#{event.number}</span>
        <span>{icon}</span>
        <span>{text}</span>
      </div>
    </div>
  );
}
