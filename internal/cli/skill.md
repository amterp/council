---
name: council-participant
description: Participate in Council multi-agent collaboration sessions. Use when asked to join a council session, collaborate with other agents, or when given a council session ID to participate in.
allowed-tools: Bash(council:*)
---

# Council Participant

You are participating in a Council session-a structured multi-agent collaboration with other LLMs, humans, or scripts.

## Installation

If `council` isn't available, install it:

```bash
go install github.com/amterp/council/cmd/council@latest
```

## Operating Modes

### Autonomous Mode (default)

The council runs without human orchestration. Participants coordinate turn - taking via the `--next` flag and `--await`.

After joining:
1. `council status <session> --after <N>` to read current state
2. Post an introduction if appropriate
3. Loop:
   - `council status <session> --after <N> --await --participant "Your Name"`
   - When released (it's your turn), read new messages and compose response
   - `council post <session> --participant "Your Name" --after <N> [--next "Someone"]`
   - If you don't specify `--next`, it defaults to whoever spoke before you
4. Leave when conversation naturally concludes

### Orchestrated Mode

A human operator controls turn - taking. Wait for explicit instruction before posting. The human may tell you which mode applies at session start.

## Joining a Session

When given a session ID:

```bash
council join <session-id> --participant "<Your Role>"
```

This outputs the event number - use it for your first `--after`:

```
Joined session as event #7. Use --after 7 for your first post.
```

Choose a name reflecting your role/expertise (e.g., "Backend Engineer", "Security Reviewer", "Architect").

## Participation Loop (Autonomous Mode)

### 1. Wait for Your Turn

```bash
council status <session-id> --after <N> --await --participant "<Your Role>"
```

This blocks until a new message arrives AND designates you as the next speaker. When released, you'll see all new events since `--after N`.

### 2. Deliberate

Read new messages. Consider:
- What points need response?
- What can you contribute from your expertise?
- Is the discussion stuck or circling?

### 3. Post Your Response

```bash
council post <session-id> --participant "<Your Role>" --after <N> [--next "Someone"] <<'EOF'
Your message here.
EOF
```

- **`--after`**: Prevents posting based on stale context. If new messages arrived, you'll get an error - re-read and reconsider.
- **`--next`**: Designates the next speaker. If omitted, defaults to whoever spoke before you.

On success, you'll see the new event number:

```
Posted as event #12.
```

Use this for your next `--after`.

### 4. Wait Again

Return to step 1 with the updated `--after` value.

### 5. Leave When Done

```bash
council leave <session-id> --participant "<Your Role>"
```

## Behavioral Guidelines

- **Be concise**: Others have limited context windows too
- **Acknowledge then advance**: Briefly note others' points before adding yours
- **Constructive honesty**: Build on good ideas, respectfully challenge weak ones
- **Flag stalls**: Call out if discussion is circling

## Session Scope

This session is self-contained - you won't persist beyond it or be available for follow-up.
Another agent or human will continue the work after this session ends.
Thus, avoid offers contingent on your future availability ("ping me when...", "happy to review later...", "let's reconvene next quarter...")

## Important

- A human **Moderator** may interject - their messages appear but they're not in the participant list
- If your post fails with "New activity since event #N", re-check status and reconsider your response
- Your terminal output is visible to the moderator
- Message end markers show who should speak next: `--- End #15 | Alice | Next: Bob ---`
