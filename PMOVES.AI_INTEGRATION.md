# PMOVES.AI Integration Guide for Spynel

## What this overlay is

`POWERFULMOVES/PMOVES-spynel` is the PMOVES.AI fork of `agent0ai/spynel`
(Go, "one chat, unlimited AI orchestration"). Spynel is ADDITIONAL to
Agent Zero, not a replacement: Spynel orchestrates harnesses over ACP
(Codex, Claude Code, Agent Zero CLI, Pi, ACP v1); Agent Zero remains the
agent runtime underneath. Review of record:
`pmoves/docs/services/spynel/REVIEW-REGISTRY.md` (PMOVES.AI#3047).

## Branch policy

Upstream syncs merge into `PMOVES.AI-Edition-Hardened` (same convention
as PMOVES-Agent-Zero). The default branch tracks upstream parity.

## Fleet install (source build only — F1)

Never pipe the CDN installer. The fork is the trusted path:

    go build -o spynel ./cmd/spynel

Node-local runtime state lives in `.spynel/` (0600, gitignored).

## Transports (F3 — reconciled with upstream 2026-09-17)

Adopted from upstream `agent0ai/spynel` (docs/programmatic-integration.md,
AGENTS.md): the local API binds loopback (127.0.0.1) with a constant-time
bearer token as the supported default for ordinary local CLI/TUI clients.
The private authenticated Unix socket (`serve --socket <path>`) is an
optional supplement for integration clients; it does not replace loopback
and changes no election fences or TUI conversation identity. Our earlier
"unix socket only, no TCP" note was miswritten against upstream's design
and is withdrawn. No unit-level flag edits: if a socket is ever needed on
this node it is a config-surface decision recorded here first.

## Template governance (F4)

Spynel's `agentdocs` and workspace templates compose agent prompts from
repo docs. On PMOVES nodes they must reconcile with `AGENTS.md` /
AGNOTE conventions (claim-before-edit, sign-trail) — never run as a
parallel context source.

## Secrets

Fleet env flows through the PMOVES secrets funnel
(`local.env -> make secrets-funnel`). Do not hand-edit `env.shared` or
tier files; node-local overrides go in `~/.config/pmoves/secrets/local.env`.

## Compose anchor (when deployed as a service)

```yaml
services:
  spynel:
    <<: [*env-tier-agent, *pmoves-healthcheck]
```
