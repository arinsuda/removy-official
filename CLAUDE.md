@AGENTS.md

# Tech Stack
Go 1.26 + Fiber v3 + GORM (server) · Vue 3 + TypeScript + Vite + Tailwind + Pinia + TanStack Query (client) · PostgreSQL 16 · MinIO · Brevo · TMDB · Socket.io.
Layout, commands, and invariants are in AGENTS.md. Don't duplicate them here.

# Coding Rules
- Client: use TypeScript in every file (`<script setup lang="ts">`). Server is Go.
- Avoid `any` unless it is really necessary.
- Keep components clearly separated and easy to read. Don't keep growing files that are already huge; split new parts into components.
- Write code that is concise and readable, but not so short it gets cryptic.
- Don't delete existing code unless you understand what it does.
- New APIs follow RESTful principles: plural nouns, HTTP verbs that match the action, correct status codes, and the response shape described in AGENTS.md.

# Workflow
- Reply to the user in Thai; keep code, comments, and commit messages in English.
- Plan before changes that touch more than 3 files or touch auth, account linking, privacy, or migrations. Ask the user when requirements are unclear.
- Run the checks before saying a task is done (see the verifier agent). Report results truthfully.
- Don't commit or push unless the user asks. Commit style: `feat:` / `fix:` / `refactor:` / `docs:`.
- After changing structure or conventions, update AGENTS.md so the next agent has an accurate map.

# Agent Orchestration (token-efficient dispatch)
The user authorizes delegation. The main session is the **orchestrator**: it decides, briefs, and integrates, and it delegates heavy reading and writing to subagents in `.claude/agents/`.
**Dispatch automatically.** For every request, classify it with the rules below and spawn the matching agent(s) without asking permission. Announce the choice in one short line, e.g. `→ scout (Haiku): หาไฟล์ที่เกี่ยวกับ watch log`. The only exception is `escalation`, which needs a one-line reason first. If the user names an agent, use that one. If an agent's file doesn't exist in `.claude/agents/`, do the work in the main session and tell the user.

## Who does what
| Agent | Model | Cost (in/out per MTok) | Use for |
|---|---|---|---|
| (main session) | your current model | — | Decisions, small edits (≤2 files, location known), talking to the user |
| `scout` | Haiku 5.5, low | $0.10 / $0.50 | Finding code, tracing call paths, summarizing files. Read-only |
| `verifier` | Haiku 5.5, low | $0.10 / $0.50 | Running vet/test/type-check, reporting only failures |
| `implementer` | Sonnet 5.5, medium | $2 / $10 | Writing code from a clear spec, plus verification |
| `architect` | Opus 5.5, high | $4 / $20 | Plans for ambiguous, cross-domain, or high-risk work |
| `reviewer` | Opus 5.5, high | $4 / $20 | Reviewing diffs touching auth, account linking, privacy, admin, migrations |
| `escalation` | Fable 5.1, high | $10 / $50 | Last resort only (see the rules below) |

## Decision rules
1. **Do it yourself** when the task is small and the location is known. Spawning has a cold-start cost; a subagent re-reads context.
2. **Don't know where the code is** → `scout` first. Never read big files (UserBmol, BmolSection, TVDetailView, MovieDetailView) in the main session.
3. **Spec is clear** → `implementer`. **Spec unclear, cross-domain, or a migration/auth/linking/privacy change** → `architect`, then `implementer`.
4. **Changed auth, account linking, privacy, admin, or migrations** → always run `reviewer` before reporting done. Other changes → `verifier` is enough.
5. **Escalate to `escalation` (Fable) only when** an Opus-level attempt failed twice, or the problem is a race/deadlock/data corruption. State the reason to the user in one line before spawning.
6. **Independent subtasks** (e.g. server + client of one feature after the API contract is fixed) → spawn them in parallel in one message.
7. Prefer the cheaper agent when in doubt; upgrade only when its output is insufficient.

## Brief template (keep under 15 lines)
```
Goal: <one sentence>
Context: <relevant facts already known — file:line, decisions made>
Scope: <files/dirs allowed to touch>  Out of scope: <...>
Done when: <acceptance criteria + check command>
Return: <what to report, e.g. "files changed + test result">
```
Pass known facts in the brief so the agent doesn't re-search them. Ask for compact output (`file:line`, no code dumps).
After an agent returns, relay the essentials to the user. Don't paste its whole report.
