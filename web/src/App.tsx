import { useSession } from './hooks/useSession';
import { useTheme } from './hooks/useTheme';
import { Header } from './components/Header';
import { MessageList } from './components/MessageList';
import { ComposeBox } from './components/ComposeBox';

function App() {
  const sessionId = new URLSearchParams(window.location.search).get('session') || '';
  const { events, participants, eventCount, loading, error, refetch } = useSession(sessionId);
  const { theme, setTheme } = useTheme();

  if (!sessionId) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50 p-4 dark:bg-gray-950">
        <div className="text-center">
          <h1 className="mb-2 text-xl font-semibold text-gray-900 dark:text-gray-100">No Session</h1>
          <p className="text-gray-600 dark:text-gray-400">
            Add <code className="rounded bg-gray-200 px-1 dark:bg-gray-700">?session=your-session-id</code> to the URL.
          </p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="text-gray-600 dark:text-gray-400">Loading session...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50 p-4 dark:bg-gray-950">
        <div className="text-center">
          <h1 className="mb-2 text-xl font-semibold text-red-600 dark:text-red-400">Error</h1>
          <p className="text-gray-600 dark:text-gray-400">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen flex-col bg-gray-50 dark:bg-gray-950">
      <Header
        sessionId={sessionId}
        participants={participants}
        theme={theme}
        onThemeChange={setTheme}
      />
      <MessageList events={events} />
      <ComposeBox
        sessionId={sessionId}
        participants={participants}
        eventCount={eventCount}
        onPostSuccess={refetch}
      />
    </div>
  );
}

export default App;
