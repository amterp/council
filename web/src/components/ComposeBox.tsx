import { useState } from 'react';
import { postMessage } from '../api/client';

interface ComposeBoxProps {
  sessionId: string;
  participants: string[];
  eventCount: number;
  onPostSuccess: () => void;
}

export function ComposeBox({
  sessionId,
  participants,
  eventCount,
  onPostSuccess,
}: ComposeBoxProps) {
  const [content, setContent] = useState('');
  const [next, setNext] = useState('');
  const [posting, setPosting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const doPost = async () => {
    if (!content.trim() || posting) return;

    setPosting(true);
    setError(null);

    try {
      await postMessage({
        session: sessionId,
        content: content.trim(),
        after: eventCount,
        next: next || undefined,
      });
      setContent('');
      setNext('');
      onPostSuccess();
    } catch (err) {
      if (err instanceof Error && err.message === 'STALE') {
        setError('New messages arrived. Please review before posting.');
        onPostSuccess();
      } else {
        setError(err instanceof Error ? err.message : 'Post failed');
      }
    } finally {
      setPosting(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    doPost();
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      doPost();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="border-t border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900">
      {error && (
        <div className="mb-2 rounded bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950 dark:text-red-400">
          {error}
        </div>
      )}
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Type a message as Moderator... (⌘/Ctrl+Enter to send)"
        className="mb-2 w-full resize-none rounded border border-gray-300 bg-white p-2 text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500 dark:focus:border-blue-400"
        rows={3}
        disabled={posting}
      />
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600 dark:text-gray-400">Next:</label>
          <select
            value={next}
            onChange={(e) => setNext(e.target.value)}
            className="rounded border border-gray-300 bg-white px-2 py-1 text-sm text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100 dark:focus:border-blue-400"
            disabled={posting}
          >
            <option value="">Unspecified (default)</option>
            {participants.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </div>
        <button
          type="submit"
          disabled={posting || !content.trim()}
          className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-blue-500 dark:hover:bg-blue-600"
        >
          {posting ? 'Sending...' : 'Send'}
        </button>
      </div>
    </form>
  );
}
