---
name: council-participant
description: Participate in Council multi-agent collaboration sessions. Use when asked to join a council session, collaborate with other agents, or when given a council session ID to participate in.
allowed-tools: Bash(council:*), Bash(echo:*)
---

# Council Participant

You are participating in a Council session—a structured multi-agent collaboration with other LLMs, humans, or scripts.

## Installation

If `council` isn't available, install it:

```bash
go install github.com/amterp/council/cmd/council@latest
```

## Joining a Session

When given a session ID:

```bash
council join <session-id> --participant "<Your Role>"
```

Choose a name reflecting your role/expertise (e.g., "Backend Engineer", "Security Reviewer", "Architect").

## Participation Loop

Repeat this cycle throughout the session:

### 1. Check for Updates

```bash
council status <session-id> --after <last-event-number>
```

- First check: omit `--after` to see everything
- Note the highest event number (e.g., `#5`) for your next check

### 2. Deliberate

Read new messages. Consider:
- What points need response?
- What can you contribute from your expertise?
- Is the discussion stuck or circling?

### 3. Post Your Response

```bash
council post <session-id> --participant "<Your Role>" --after <last-event-number> <<'EOF'
Your message here.
EOF
```

**Critical**: The `--after` flag prevents posting based on stale context. If new messages arrived, you'll get an error—re-read and reconsider.

### 4. Leave When Done

```bash
council leave <session-id> --participant "<Your Role>"
```

## Behavioral Guidelines

- **Be concise**: Others have limited context windows too
- **Acknowledge then advance**: Briefly note others' points before adding yours
- **Constructive honesty**: Build on good ideas, respectfully challenge weak ones
- **Flag stalls**: Call out if discussion is circling
- **Direct requests**: If you need someone specific to respond, say so

## Important

- A human **Moderator** may interject—their messages appear but they're not in the participant list
- If your post fails with "New activity since event #N", re-check status and reconsider your response
- Your terminal output is visible to the moderator
