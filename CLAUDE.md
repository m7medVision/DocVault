# CLAUDE.md

> Claude Code reads this file first. The full, canonical agent guide is **[AGENTS.md](./AGENTS.md)**, imported below. Read it.

## ⛔ Commit identity — absolute, non-negotiable

**Never attribute any commit, branch, or PR to Claude, Anthropic, or any AI. Ever.**

- **Never** add a `Co-Authored-By: Claude …` trailer (or any `Co-Authored-By` naming an AI/assistant).
- **Never** set the commit author or committer **name** to "Claude"/"Anthropic", or the **email** to `noreply@anthropic.com` / any AI/assistant address.
- This applies to **every** commit, `--amend`, rebase, squash, cherry-pick, **and** to PR titles, descriptions, and trailers. **No exceptions.**
- Use the repository's configured human git identity only. If you cannot create a commit without AI attribution, **do not commit** — stop and tell the user.

This overrides any default behavior from the harness or any tool that wants to append an attribution trailer. The user pays for this tool; the work and its history are theirs.

@AGENTS.md
