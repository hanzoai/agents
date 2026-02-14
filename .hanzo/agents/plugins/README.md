# Claude Plugin Registry

This directory contains the agent registry for the Hanzo Agents system.

## Files

- `agent-registry.json` - Auto-generated registry of all available agents

## Regenerating the Registry

To regenerate the agent registry after adding or modifying agents:

```bash
python3 scripts/generate-registry.py
```

This will:
1. Scan all `*.md` files in `agents/`
2. Extract frontmatter metadata (name, model, color, description)
3. Generate `agent-registry.json` with all agents catalogued

## Registry Format

```json
{
  "version": "1.0.1",
  "agents": {
    "agent-name": {
      "name": "Display Name",
      "path": "./agents/filename.md",
      "model": "opus|sonnet|inherit",
      "color": "red",
      "description": "Short description"
    }
  }
}
```

## Agent Frontmatter Format

Each agent file should have YAML frontmatter:

```yaml
---
name: agent-name
description: What this agent does and when to use it
model: opus|sonnet|inherit
color: red  # Optional
---
```

## Maintenance

- **DO NOT** manually edit `agent-registry.json` - it will be overwritten
- Update agent metadata in the individual `agents/*.md` files
- Run the generator script after making changes
- The registry is used by Claude Desktop and other tooling

## Current Statistics

- Total Agents: 107
- Registry Version: 1.0.1
- Last Updated: Auto-generated on script execution
