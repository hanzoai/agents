# Hanzo MCP Setup Guide

Complete guide for installing and configuring Hanzo Model Context Protocol with Claude Code.

## Prerequisites

- Python 3.9+ or Node.js 18+
- Claude Code installed
- Git (for repository installation)

## Installation Methods

### Method 1: Python with UV (Recommended)

**Fastest and most reliable:**

```bash
# Install uvx if not already installed
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install hanzo-mcp
uvx hanzo-mcp

# Verify installation
uvx hanzo-mcp --version
```

### Method 2: Python with pip

```bash
# Install globally
pip install hanzo-mcp

# Or with pipx for isolated install
pipx install hanzo-mcp

# Verify installation
hanzo-mcp --version
```

### Method 3: Node.js/TypeScript

```bash
# Install globally
npm install -g @hanzo/mcp

# Or with pnpm
pnpm add -g @hanzo/mcp

# Verify installation
hanzo-mcp --version
```

### Method 4: Rust (Most Performant)

```bash
# Install from crates.io
cargo install hanzo-mcp

# Verify installation
hanzo-mcp --version
```

## Claude Desktop Configuration

### Automatic Installation

```bash
# Install MCP server to Claude Desktop automatically
hanzo-mcp install-desktop

# This configures Claude Desktop to use hanzo-mcp
# Location: ~/.config/Claude/claude_desktop_config.json (Linux/macOS)
#           %APPDATA%/Claude/claude_desktop_config.json (Windows)
```

### Manual Configuration

Edit Claude Desktop config file:

**macOS/Linux**: `~/.config/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "hanzo": {
      "command": "uvx",
      "args": ["hanzo-mcp", "serve"],
      "env": {
        "HANZO_MCP_LOG_LEVEL": "info"
      }
    }
  }
}
```

**Alternative configurations:**

```json
// Using pip install
{
  "mcpServers": {
    "hanzo": {
      "command": "hanzo-mcp",
      "args": ["serve"]
    }
  }
}

// Using Node.js
{
  "mcpServers": {
    "hanzo": {
      "command": "npx",
      "args": ["@hanzo/mcp", "serve"]
    }
  }
}

// Using Rust
{
  "mcpServers": {
    "hanzo": {
      "command": "hanzo-mcp",
      "args": ["serve", "--rust"]
    }
  }
}
```

## Verify Installation

### 1. Check MCP Server

```bash
# Start MCP server manually to test
hanzo-mcp serve --port 3000

# In another terminal, test connection
curl http://localhost:3000/health

# Expected response:
# {"status": "ok", "version": "1.0.0", "tools": 42}
```

### 2. Test in Claude Desktop

1. Restart Claude Desktop
2. Open a new conversation
3. Type: "List available MCP tools"
4. You should see hanzo-mcp tools listed

### 3. Test Tool Invocation

```
You: "Use hanzo-mcp to read the file at /path/to/file.py"
Claude: [Uses MCP read tool and shows file contents]

You: "Search for TODO comments in my project"
Claude: [Uses MCP search tool and shows results]
```

## Configuration Options

### Environment Variables

```bash
# API Keys for agent tools
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...

# MCP server settings
export HANZO_MCP_PORT=3000
export HANZO_MCP_HOST=localhost
export HANZO_MCP_LOG_LEVEL=info  # debug, info, warn, error

# Tool permissions
export HANZO_MCP_PERMISSION_MODE=review  # review, auto_approve, auto_deny

# Search settings
export HANZO_MCP_SEARCH_MAX_RESULTS=100
export HANZO_MCP_SEARCH_IGNORE=node_modules,*.pyc,.git

# File operations
export HANZO_MCP_MAX_FILE_SIZE=10485760  # 10MB
export HANZO_MCP_TIMEOUT=120000  # 2 minutes
```

### YAML Configuration File

Create `~/.hanzo/mcp/config.yaml`:

```yaml
server:
  host: localhost
  port: 3000
  log_level: info

tools:
  filesystem:
    enabled: true
    max_file_size: 10MB
    allowed_paths:
      - /home/user/projects
      - /workspace
    blocked_paths:
      - /etc
      - /var
      - /sys

  shell:
    enabled: true
    timeout: 120000  # 2 minutes
    auto_background: true
    allowed_commands:
      - npm
      - pnpm
      - python
      - pytest
      - git
    blocked_commands:
      - rm -rf /
      - sudo
      - chmod 777

  search:
    enabled: true
    max_results: 100
    ignore_patterns:
      - node_modules
      - "*.pyc"
      - .git
      - .venv
      - __pycache__
    enable_vector: true
    enable_ast: true
    enable_symbol: true

  agent:
    enabled: true
    max_concurrent: 5
    models:
      - claude-opus-4
      - claude-sonnet-4

permissions:
  mode: review  # review, auto_approve, auto_deny
  whitelist:
    - read
    - grep
    - search
    - ast
    - directory_tree
  require_approval:
    - write
    - edit
    - multi_edit
    - bash
    - shell
  blacklist:
    - rm
    - sudo
    - chmod

monitoring:
  enabled: true
  metrics_port: 9090
  log_dir: ~/.hanzo/mcp/logs
```

## Claude Code Integration

### Install Agents Repository

```bash
# Method 1: Via Claude Code plugin system
# In Claude Code, run:
/plugin install hanzo-agents

# Method 2: Manual clone
cd ~/.claude
git clone https://github.com/hanzoai/agents.git
```

### Verify Agents Available

In Claude Code:
```
You: "List available agents"
Claude: [Shows all installed agents including hanzo, dev, cto, etc.]

You: "Use dev agent to implement a rate limiter"
Claude: [Activates dev agent with first-principles approach]
```

## Troubleshooting

### Issue: MCP Server Not Starting

**Check logs:**
```bash
cat ~/.hanzo/mcp/logs/server.log
```

**Common causes:**
- Port already in use → Change port in config
- Missing dependencies → Reinstall hanzo-mcp
- Permission issues → Check file permissions

**Solution:**
```bash
# Kill existing process
pkill -f hanzo-mcp

# Restart with different port
hanzo-mcp serve --port 3001
```

### Issue: Tools Not Appearing in Claude Desktop

**Verify config:**
```bash
# Check config file exists
cat ~/.config/Claude/claude_desktop_config.json

# Restart Claude Desktop
killall Claude
open -a Claude
```

**Check command path:**
```bash
# Ensure command is in PATH
which hanzo-mcp
which uvx

# If not found, use absolute path in config
{
  "mcpServers": {
    "hanzo": {
      "command": "/Users/yourname/.local/bin/hanzo-mcp",
      "args": ["serve"]
    }
  }
}
```

### Issue: Permission Denied Errors

**Adjust permission mode:**
```bash
# Temporary: Set to auto-approve
export HANZO_MCP_PERMISSION_MODE=auto_approve

# Or in config.yaml
permissions:
  mode: auto_approve
```

**Check allowed paths:**
```yaml
tools:
  filesystem:
    allowed_paths:
      - /your/project/path  # Add your project directory
```

### Issue: Agent Tools Not Working

**Check API keys:**
```bash
# Verify keys are set
echo $ANTHROPIC_API_KEY
echo $OPENAI_API_KEY

# Add to shell profile
echo 'export ANTHROPIC_API_KEY=sk-ant-...' >> ~/.zshrc
source ~/.zshrc
```

**Test agent invocation:**
```bash
# Test dispatch_agent directly
hanzo-mcp test dispatch-agent "List files in current directory"
```

### Issue: Search Tools Slow

**Optimize search configuration:**
```yaml
search:
  max_results: 50  # Reduce from 100
  ignore_patterns:
    - node_modules
    - dist
    - build
    - "*.min.js"
  enable_vector: false  # Disable if not needed
```

### Issue: High Memory Usage

**Reduce concurrent operations:**
```yaml
agent:
  max_concurrent: 3  # Reduce from 5

tools:
  batch:
    max_batch_size: 5  # Reduce from 10
```

## Advanced Configuration

### Custom Tool Development

```python
# ~/.hanzo/mcp/custom_tools.py
from hanzo_mcp import Tool, ToolContext

class MyCustomTool(Tool):
    name = "my_tool"
    description = "Custom tool for specific workflow"

    async def call(self, ctx: ToolContext, **params):
        # Implementation
        return "Result"

# Register in config.yaml
tools:
  custom:
    enabled: true
    module: ~/.hanzo/mcp/custom_tools.py
    tools:
      - MyCustomTool
```

### Multi-Server Setup

```json
{
  "mcpServers": {
    "hanzo-dev": {
      "command": "uvx",
      "args": ["hanzo-mcp", "serve", "--config", "~/.hanzo/mcp/dev.yaml"]
    },
    "hanzo-prod": {
      "command": "uvx",
      "args": ["hanzo-mcp", "serve", "--config", "~/.hanzo/mcp/prod.yaml"]
    }
  }
}
```

### Performance Tuning

```yaml
# High-performance configuration
server:
  workers: 4  # CPU cores
  max_connections: 100

tools:
  filesystem:
    cache_enabled: true
    cache_ttl: 300  # 5 minutes

  search:
    parallel_search: true
    max_workers: 8

  agent:
    pool_size: 10
    keepalive: true
```

## Security Best Practices

1. **Use Review Mode in Production**
   ```yaml
   permissions:
     mode: review  # Always review destructive operations
   ```

2. **Restrict File Access**
   ```yaml
   tools:
     filesystem:
       allowed_paths:
         - /workspace
         - /home/user/projects
       blocked_paths:
         - /etc
         - /root
         - ~/.ssh
   ```

3. **Limit Shell Commands**
   ```yaml
   tools:
     shell:
       blocked_commands:
         - rm -rf
         - sudo
         - chmod 777
         - dd
   ```

4. **API Key Security**
   ```bash
   # Never commit API keys
   # Use environment variables or secret management

   # Rotate keys regularly
   # Use different keys per environment (dev/prod)
   ```

5. **Audit Logging**
   ```yaml
   monitoring:
     audit_log: true
     log_dir: ~/.hanzo/mcp/audit
     retention_days: 90
   ```

## Integration with Hanzo Ecosystem

### Hanzo Cloud Platform

```yaml
# Connect to Hanzo Cloud
cloud:
  enabled: true
  endpoint: https://api.hanzo.ai
  api_key: ${HANZO_API_KEY}
  project_id: ${HANZO_PROJECT_ID}
```

### Hanzo Analytics

```yaml
# Enable analytics tracking
analytics:
  enabled: true
  endpoint: https://analytics.hanzo.ai
  track_tool_usage: true
  track_agent_performance: true
```

### Hanzo LLM Gateway

```yaml
# Use Hanzo LLM Gateway for all AI operations
llm:
  gateway: https://llm.hanzo.ai
  api_key: ${HANZO_LLM_API_KEY}
  default_model: claude-opus-4
  fallback_model: claude-sonnet-4
```

## Health Checks

### Automated Health Monitoring

```bash
# Add to crontab for monitoring
*/5 * * * * curl -s http://localhost:3000/health || echo "MCP server down!" | mail -s "Alert: MCP Down" ops@company.com
```

### Prometheus Metrics

```yaml
# Enable Prometheus metrics
monitoring:
  prometheus:
    enabled: true
    port: 9090
    metrics:
      - tool_invocations_total
      - tool_latency_seconds
      - agent_executions_total
      - cache_hit_rate
```

## Support

**Documentation:**
- [Hanzo MCP API Docs](https://docs.hanzo.ai/mcp)
- [Claude Code Integration](https://docs.anthropic.com/claude-code)
- [Troubleshooting Guide](https://docs.hanzo.ai/mcp/troubleshooting)

**Community:**
- GitHub Issues: https://github.com/hanzoai/hanzo-mcp/issues
- Discord: https://discord.gg/hanzo-ai
- Email: support@hanzo.ai

**Updates:**
```bash
# Update hanzo-mcp
uvx --upgrade hanzo-mcp

# Or with pip
pip install --upgrade hanzo-mcp
```

---

**Last Updated**: 2025-01-09
**Version**: 1.0.0
