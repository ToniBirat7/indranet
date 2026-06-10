# .pulse/ — managed by the `pulse` skill

This folder is the **canonical state** of this project for the Pulse dashboard.

- `state.json` — structured status, stack, branches, metrics. **Machine fields are
  regenerated on `pulse sync`** — don't hand-edit them; put overrides in `state.json` → `human{}`.
- `tasks.json` — kanban tasks (todo/doing/done). Set `"locked": true` to freeze a task.
- `timeline.jsonl` — append-only history. Never rewritten.
- `overview.md` — yours. Prose vision + architecture.
- `pipeline.md` — yours, except between the `pulse:auto` markers.
- `wiki.json` — read-only pointer to your `wiki/` research. The skill never writes into `wiki/`.
- `.pulsemeta.json` — machine bookkeeping (gitignored).

Refresh: run `pulse sync` (or it's stamped dirty automatically when a Claude session ends).
