import type { Theme } from '../hooks/useTheme';

interface HeaderProps {
  sessionId: string;
  participants: string[];
  theme: Theme;
  onThemeChange: (theme: Theme) => void;
}

export function Header({ sessionId, participants, theme, onThemeChange }: HeaderProps) {
  return (
    <div className="border-b border-gray-200 bg-white px-4 py-3 dark:border-gray-700 dark:bg-gray-900">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Council: {sessionId}
          </h1>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Participants: {participants.length > 0 ? participants.join(', ') : 'None yet'}
          </p>
        </div>
        <div className="flex items-center gap-1">
          <ThemeButton
            icon="☀️"
            label="Light"
            active={theme === 'light'}
            onClick={() => onThemeChange('light')}
          />
          <ThemeButton
            icon="🌙"
            label="Dark"
            active={theme === 'dark'}
            onClick={() => onThemeChange('dark')}
          />
          <ThemeButton
            icon="💻"
            label="System"
            active={theme === 'system'}
            onClick={() => onThemeChange('system')}
          />
        </div>
      </div>
    </div>
  );
}

interface ThemeButtonProps {
  icon: string;
  label: string;
  active: boolean;
  onClick: () => void;
}

function ThemeButton({ icon, label, active, onClick }: ThemeButtonProps) {
  return (
    <button
      onClick={onClick}
      title={label}
      className={`rounded px-2 py-1 text-sm transition-colors ${
        active
          ? 'bg-gray-200 dark:bg-gray-700'
          : 'hover:bg-gray-100 dark:hover:bg-gray-800'
      }`}
    >
      {icon}
    </button>
  );
}
