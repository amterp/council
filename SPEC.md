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
~/.council/sessions/<session-id>/events.jsonl
```

Each session is a directory containing `events.jsonl`.

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
| `message` | `participant`, `content`, `next` | A contribution to the discussion. `next` designates who should speak next. |

**Example session file:**
```jsonl
{"type": "session_created", "id": "hopeful-coral-tiger", "timestamp_millis": 1705312200000}
{"type": "joined", "participant": "Engineer", "timestamp_millis": 1705312260000}
{"type": "joined", "participant": "Architect", "timestamp_millis": 1705312265000}
{"type": "message", "participant": "Engineer", "content": "I think we need OAuth2.", "next": "Architect", "timestamp_millis": 1705312290000}
{"type": "message", "participant": "Architect", "content": "Agreed. Let's design the flow.", "next": "Engineer", "timestamp_millis": 1705312350000}
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
- **Returns the event number** for use with first `--after`

**Flags:**
- `--participant <name>` or `-p`: Provide name without interactive prompt

**Output:**
```
Joined session as event #7. Use --after 7 for your first post.
```

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
- `--await`: Block until new events arrive AND it's your turn (requires `--participant`)
- `--participant <name>` or `-p`: Your participant name (required with `--await`)
- `--timeout <seconds>`: Timeout for `--await` (default: 300)

**Await behavior:**
When `--await` is used, the command blocks until:
1. Event count exceeds `--after N`
2. The latest message's `next` field matches `--participant`

If the latest message's `next` doesn't match, the command auto-increments its internal after counter and continues waiting.

**Output format:**
```
=== Session: hopeful-coral-tiger ===
Participants: Engineer, Architect

--- #5 | Engineer Joined ---

--- #6 | Architect Joined ---

--- #7 | Engineer ---
I think we need to consider OAuth2 from the start.
Here's my reasoning...
--- End #7 | Engineer | Next: Architect ---

--- #8 | Architect ---
Agreed. Let's discuss the token flow.
--- End #8 | Architect | Next: Engineer ---
```

**Notes:**
- Messages have explicit start and end markers
- End markers include the author and next speaker: `--- End #N | Author | Next: Speaker ---`
- Join/leave events shown inline as single-line entries
- No timestamps in output (reduces noise for LLM context)
- Event numbers shown as `#N`

---

### `council post <session-id>`
Posts a message to the session.

- Content via stdin or `--file`
- Requires participant name via `--participant`
- Requires `--after N` for optimistic locking (safety measure)
- **Returns the new event number**

**Flags:**
- `--participant <name>` or `-p`: Required. Who is posting.
- `--file <path>` or `-f`: Read content from file instead of stdin.
- `--after N`: Required. Only post if latest event is exactly N. Fail otherwise.
- `--next <name>` or `-n`: Optional. Designate the next speaker.

**`--next` defaulting:**
If `--next` is not provided, it defaults to:
1. Previous speaker (author of the message before this one)
2. If none: random active participant (excluding self)
3. If none: "Moderator"

The `--next` value is validated: must be an active participant or "Moderator".

**Output:**
```
Posted as event #12.
```

**Errors:**
- Participant not in session (hasn't joined)
- `--after` mismatch: "New activity since event #N. Re-check with 'council status <id> --after N'"
- Invalid `--next`: "<name> is not an active participant or 'Moderator'. Cannot use as --next."

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

See the separate `SKILL.md` file for the full participant instructions. Key concepts:

- **Session Scope**: Sessions are self-contained. Focus on durable artifacts, not personal follow-up commitments.
- **Operating Modes**: Autonomous (default, uses `--await`) vs Orchestrated (human controls turn-taking).
- **Turn-Taking**: Use `--next` to designate who speaks next. Use `--await` to wait for your turn.
- **Event Numbers**: `join` and `post` return event numbers for your next `--after`.

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
