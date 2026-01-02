# Council - Multi-Agent Collaboration Tool Specification

## Overview

Council is a CLI tool for running collaborative sessions between multiple participants. Participants can be LLM instances, humans, scripts - anything that can execute shell commands.

**Core philosophy**: The session file is the single source of truth. The CLI is stateless. Any frontend can read the session file and present it however it wants.

---

## Design Principles

- **Stateless CLI**: All state lives in the session file. The CLI reads/writes but holds nothing.
- **File-based**: Sessions stored as JSONL in `~/.council/sessions/<id>.jsonl`
- **Frontend-agnostic**: CLI manages state. Any frontend (TUI, web, etc.) can read/display.
- **Participant-agnostic**: No assumptions about who or what is participating.
- **Helpful errors**: Error messages should guide participants toward resolution.

---

## Storage

### Location
```
~/.council/sessions/<session-id>.jsonl
```

### Format
JSONL (JSON Lines). Each line is a self-contained event. Line number = event number (1-indexed for display, 0-indexed in file).

### Schema

All events share:
```json
{"type": "<event_type>", "timestamp_millis": <epoch_millis>, ...}
```

**Event types:**

| Type | Additional Fields | Description |
|------|-------------------|-------------|
| `session_created` | `id` | First line. Created by `council new`. |
| `joined` | `participant` | A participant entered the session. |
| `left` | `participant` | A participant departed the session. |
| `message` | `participant`, `content` | A contribution to the discussion. |

**Example session file:**
```jsonl
{"type": "session_created", "id": "hopeful-coral-tiger", "timestamp_millis": 1705312200000}
{"type": "joined", "participant": "Moderator", "timestamp_millis": 1705312205000}
{"type": "message", "participant": "Moderator", "content": "Welcome everyone. Today we're designing...", "timestamp_millis": 1705312210000}
{"type": "joined", "participant": "Engineer", "timestamp_millis": 1705312260000}
{"type": "message", "participant": "Engineer", "content": "Thanks for having me.", "timestamp_millis": 1705312290000}
{"type": "left", "participant": "Engineer", "timestamp_millis": 1705316400000}
```

---

## Concurrency

### File Locking
All write operations MUST acquire an exclusive file lock before modifying the session file.

### Optimistic Locking Pattern
The `council post` command requires an `--after N` flag for optimistic concurrency:

1. Participant reads current state (no lock needed)
2. Participant deliberates (no lock held)
3. Participant calls `council post --after N` where N is the last event number they saw
4. CLI acquires lock
5. CLI checks: is current latest event number == N?
   - **Yes**: Write message, release lock, return success
   - **No**: Release lock, return error with message indicating new activity

This prevents participants from posting based on stale state without holding locks during deliberation.

### Sanity Checks After Lock
After acquiring the lock, the CLI performs validation:
- Duplicate participant name check (for `join`)
- Reserved name check ("Moderator" is reserved)
- Session file integrity

---

## CLI Commands

### `council new`
Creates a new session. Prints session ID.

- Generates ID using golang-petname (3 words, e.g., `hopeful-coral-tiger`)
- Creates session file with `session_created` event
- Does NOT auto-join any participant

**Output:** Session ID (e.g., `hopeful-coral-tiger`)

---

### `council join <session-id>`
Joins a session as a named participant.

- Prompts for participant name (or accept via flag)
- Acquires lock, checks for duplicate names, appends `joined` event
- Rejects reserved name "Moderator"

**Flags:**
- `--name <name>`: Provide name without interactive prompt

**Errors:**
- Session not found
- Name already taken
- Name is reserved ("Moderator")

---

### `council leave <session-id>`
Leaves a session.

- Requires participant name (flag or prompt)
- Appends `left` event

**Flags:**
- `--name <name>`: Participant name

---

### `council status <session-id> [--after N]`
Displays session state.

- Shows participants list (excluding Moderator)
- Shows messages (all, or only after event N if `--after` provided)
- Human-readable format with clear demarcation

**Flags:**
- `--after N`: Only show events after event number N

**Output format:**
```
=== Session: hopeful-coral-tiger ===
Participants: Moderator, Engineer

--- #5 | Moderator ---
Welcome everyone. Today we're designing the new auth system.
Let's start with requirements.
--- End #5 ---

--- #6 | Engineer Joined ---

--- #7 | Engineer ---
I think we need to consider OAuth2 from the start.
Here's my reasoning...
--- End #7 ---

--- #8 | Engineer Left ---
```

**Notes:**
- Messages have explicit start and end markers
- Join/leave events shown inline as single-line entries
- No timestamps in output (reduces noise for LLM context)
- Event numbers shown as `#N`

---

### `council post <session-id>`
Posts a message to the session.

- Content via stdin or `--file`
- Requires participant name via `--participant`
- Requires `--after N` for optimistic locking (safety measure)

**Flags:**
- `--participant <name>`: Required. Who is posting.
- `--file <path>`: Read content from file instead of stdin.
- `--after N`: Required. Only post if latest event is exactly N. Fail otherwise.

**Errors:**
- Participant not in session (hasn't joined)
- `--after` mismatch: "New activity since event #N. Re-check with 'council status <id> --after N'"

---

### `council watch <session-id>`
TUI frontend for watching and participating.

- Chat-style layout: messages scroll above, input box below
- Starts as spectator (no join event)
- Typing and submitting posts as "Moderator" (invisible participant)
- Polling for updates (reasonable interval for snappy feel, ~500ms-1s)

**Moderator behavior:**
- "Moderator" is a reserved, invisible participant
- Does not appear in participants list
- No join/leave events for Moderator
- Multiple `watch` instances all post as "Moderator"

---

## Session IDs

Generated using [golang-petname](https://github.com/dustinkirkland/golang-petname):
- 3 words (e.g., `hopeful-coral-tiger`)
- Hyphen-separated
- No user-provided override for MVP

---

## Error Handling Philosophy

All errors should be helpful and actionable:

| Scenario | Error Message |
|----------|---------------|
| Session not found | `Session 'xyz' not found. Run 'council new' to create a session.` |
| Name taken | `Participant 'Engineer' already exists in this session. Choose a different name.` |
| Reserved name | `'Moderator' is a reserved name. Choose a different name.` |
| Stale post | `New activity since event #5. Re-read with 'council status <id> --after 5' before posting.` |
| Not a participant | `You must join the session before posting. Run 'council join <id>'.` |

---

## SKILL.md (for LLM Participants)

```markdown
# Council - Multi-Agent Collaboration

Participate in collaborative sessions with other participants and a moderator.

## Joining

1. Run `council join <session-id>`
2. Review current participants
3. Enter an identifier based on your role/expertise (must be unique)

## Participating

1. Check for updates:
   ```bash
   council status <session-id> --after <last-seen-event>
   ```
   Use `--after 0` on first check (or omit to see all).

2. Read new messages and deliberate your response.

3. Post your contribution:
   ```bash
   council post <session-id> --participant "Your Name" --after <last-seen-event> <<EOF
   Your message here...
   EOF
   ```

   The `--after` flag ensures you don't post based on stale context. If new messages arrived, you'll get an error prompting you to re-read.

4. Track the latest event number for your next `--after` call.

## Norms

- **Constructive honesty**: Build on strong ideas. Respectfully challenge weak ones. Don't agree just to agree.
- **Acknowledge then advance**: Briefly acknowledge others' points before adding your own. Don't let good points get lost.
- **Be concise**: Respect others' context windows. Trust them to ask if they need more - and do the same.
- **Direct when needed**: If you want someone specific to respond next, suggest it.
- **Flag stalls**: If the discussion is circling or stuck, call it out.

## Important

- Your terminal output is visible to the moderator watching you work
- Only `council post` messages are part of the shared record
- If your post fails due to new activity, re-read and reconsider before posting
- A human "Moderator" may interject occasionally. They don't appear in the participants list but their messages are visible.
```

---

## Technology Stack

- **Language**: Go
- **TUI**: Bubble Tea (or similar Go TUI library)
- **ID Generation**: golang-petname
- **Target Platform**: macOS primary, but platform-agnostic design

---

## Out of Scope (for MVP)

- Session archival/deletion
- JSON output format (planned, not implemented)
- Web frontend
- User-provided session IDs
- Authentication/access control
