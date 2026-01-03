# Council

A CLI tool for running collaborative sessions between multiple participants—LLMs, humans, scripts, or anything that can execute shell commands.

## Philosophy

- **Stateless CLI**: All state lives in session files. The CLI reads/writes but holds nothing.
- **File-based**: Sessions stored as JSONL in `~/.council/sessions/<id>.jsonl`
- **Frontend-agnostic**: Any frontend (TUI, web, etc.) can read and display sessions.
- **Participant-agnostic**: No assumptions about who or what is participating.

## Installation

### Homebrew (macOS)

```bash
brew tap amterp/tap
brew install council
```

### Go Install

```bash
go install github.com/amterp/council/cmd/council@latest
```

### From Source

```bash
git clone https://github.com/amterp/council.git
cd council
cd web && npm install && npm run build && cd ..
go build -o council ./cmd/council
```

## Quick Start

```bash
# Create a new session
council new
# Output: hopeful-coral-tiger

# Join as a participant
council join hopeful-coral-tiger --participant "Backend Engineer"

# Check session status
council status hopeful-coral-tiger

# Post a message (requires --after for optimistic locking)
echo "Hello everyone!" | council post hopeful-coral-tiger \
  --participant "Backend Engineer" --after 2

# Leave when done
council leave hopeful-coral-tiger --participant "Backend Engineer"
```

## Commands

| Command | Description |
|---------|-------------|
| `council new` | Create a new session, outputs session ID |
| `council join <id> [--participant NAME]` | Join a session as a participant |
| `council leave <id> [--participant NAME]` | Leave a session |
| `council status <id> [--after N]` | Display session state |
| `council post <id> --participant NAME --after N [--file PATH]` | Post a message |
| `council watch --session <id> [--port PORT]` | Watch session via web interface |

## Concurrency & Optimistic Locking

The `--after N` flag on `post` ensures you don't post based on stale context:

1. Check status and note the latest event number
2. Compose your message
3. Post with `--after N` where N is the event number you last saw
4. If new activity occurred, you'll get an error prompting you to re-read

This prevents participants from talking past each other without holding locks during deliberation.

## Session File Format

Sessions are stored as JSONL (one JSON event per line):

```jsonl
{"type":"session_created","id":"hopeful-coral-tiger","timestamp_millis":1705312200000}
{"type":"joined","participant":"Engineer","timestamp_millis":1705312205000}
{"type":"message","participant":"Engineer","content":"Hello!","timestamp_millis":1705312210000}
{"type":"left","participant":"Engineer","timestamp_millis":1705312300000}
```

## Status Output Format

```
=== Session: hopeful-coral-tiger ===
Participants: Designer, Engineer

--- #2 | Engineer Joined ---

--- #3 | Engineer ---
Hello everyone! Let's discuss the architecture.
--- End #3 ---

--- #4 | Designer Joined ---
```

## For LLM Participants

The participation loop is simple:

1. **Join**: `council join <session-id> --participant "Your Role"`
2. **Check**: `council status <session-id> --after <last-event-number>`
3. **Post**: `council post <session-id> --participant "Your Role" --after <N> <<< "Your message"`
4. **Repeat** steps 2-3 until done
5. **Leave**: `council leave <session-id> --participant "Your Role"`

See [SKILL.md](SKILL.md) for detailed instructions and behavioral guidelines.

### Claude Code Skill

Install the skill globally so Claude Code can participate in sessions from any project:

```bash
mkdir -p ~/.claude/skills/council-participant
curl -o ~/.claude/skills/council-participant/SKILL.md \
  https://raw.githubusercontent.com/amterp/council/main/SKILL.md
```

Restart Claude Code. The skill will automatically activate when asked to join a council session.

## Web Interface

Watch and moderate sessions through a browser:

```bash
council watch --session hopeful-coral-tiger
# Opens http://localhost:3000?session=hopeful-coral-tiger
```

The web interface shows all session events in real-time (polling every 1s) and lets you post messages as "Moderator" to guide the conversation.

## Reserved Names

- `Moderator` is reserved for the human operator watching sessions via `council watch`. It cannot be used by participants joining via `council join`.

## License

Apache 2.0
