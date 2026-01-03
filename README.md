# Council

A CLI tool for running collaborative sessions between multiple LLM agents, with optional human moderation.

## Philosophy

- **Stateless CLI**: All state lives in session files. The CLI reads/writes but holds nothing.
- **File-based**: Sessions stored as JSONL in `~/.council/sessions/<id>/events.jsonl`
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

See [example](#claude-code-integration) of using this with Claude Code.

## Commands

| Command                                                        | Description                                           |
|----------------------------------------------------------------|-------------------------------------------------------|
| `council new`                                                  | Create a new session, outputs session ID              |
| `council join <id> [--participant NAME]`                       | Join a session as a participant                       |
| `council leave <id> [--participant NAME]`                      | Leave a session                                       |
| `council status <id> [--after N]`                              | Display session state                                 |
| `council post <id> --participant NAME --after N [--file PATH]` | Post a message                                        |
| `council watch --session <id> [--port PORT]`                   | Watch session via web interface                       |
| `council install <target>`                                     | Install integrations (e.g., `council install claude`) |

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

### Claude Code Integration

Install the council-participant skill so Claude Code can join sessions from any project:

```bash
council install claude
```

This installs the skill to `~/.claude/skills/council-participant/`. Restart Claude Code to activate.

> **Alternative**: You can manually download [SKILL.md](SKILL.md) to `~/.claude/skills/council-participant/SKILL.md`.

### Example: Multi-Agent Collaboration with Claude Code

Council shines when you have multiple Claude Code instances collaborate - each bringing different expertise
via [output-styles](https://docs.anthropic.com/en/docs/claude-code/settings#output-style), or working from different
codebases.

**The scenario**: You're designing a new syntax feature for a programming language. You want diverse perspectives:

| Terminal | Role                  | Notes                                                             |
|----------|-----------------------|-------------------------------------------------------------------|
| 1        | **Software Engineer** | Main compiler repo, general SWE output-style                      |
| 2        | **Grammar Expert**    | Grammar repo, output-style tuned for parsing/syntax               |
| 3        | **Devil's Advocate**  | Main repo, output-style customized for constructive contrarianism |
| 4        | Human moderator       | Creates session, watches via web UI                               |

For the Devil's Advocate, you might add something like this to their output-style:

> *Your role is to be a constructive contrarian. Challenge assumptions, identify blind spots, and guard against
groupthink. When everyone agrees too quickly, probe deeper. Stress-test ideas before they're committed to.*

**Step 1: Create the session** (Terminal 4)

```bash
council new                                  # -> hopeful-coral-tiger
council watch --session hopeful-coral-tiger  # opens web UI
```

**Step 2: Brief each Claude** (Terminals 1-3)

Give each Claude the task context and session ID:

> *"We're designing a new pattern-matching syntax. Explore the codebase to understand our current approach, then join
council session `hopeful-coral-tiger` to collaborate with the other participants."*

Each Claude will:

1. Explore its codebase to gather relevant context
2. Join the session with an appropriate participant name
3. Introduce itself and share relevant findings
4. Wait for its turn (via `--await`), contribute ideas, respond to others
5. Leave when the discussion concludes

**Step 3: Watch and moderate**

Follow the discussion in the web UI. As "Moderator", you can steer the conversation, add constraints, or call for
decisions.

**Why this works**:

- Each agent brings a distinct perspective shaped by its output-style and codebase context
- The Devil's Advocate ensures ideas get stress-tested before consensus
- You maintain oversight without being in the critical path
- All discussion is logged to `~/.council/sessions/<id>/events.jsonl`

## Web Interface

Watch and moderate sessions through a browser:

```bash
council watch --session hopeful-coral-tiger
# Opens http://localhost:3000?session=hopeful-coral-tiger
```

The web interface shows all session events in real-time (polling every 1s) and lets you post messages as "Moderator" to
guide the conversation.

## Reserved Names

- `Moderator` is reserved for the human operator watching sessions via `council watch`. It cannot be used by
  participants joining via `council join`.

## License

Apache 2.0
