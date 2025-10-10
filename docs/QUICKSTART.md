# Quick Start: Hanzo Agents for Claude Code

Get started with AI-powered development in 5 minutes.

## 1. Install Prerequisites (2 minutes)

### Install Hanzo MCP

**Choose one method:**

```bash
# Fastest: UV (recommended)
curl -LsSf https://astral.sh/uv/install.sh | sh
uvx hanzo-mcp

# Or: pip
pip install hanzo-mcp

# Or: npm/pnpm
npm install -g @hanzo/mcp
```

### Configure Claude Desktop

```bash
# Automatic setup
hanzo-mcp install-desktop

# Restart Claude Desktop
```

## 2. Install Core Agents (1 minute)

### Method 1: Via Claude Code Plugin

In Claude Code:
```
/plugin install hanzo-core-team
```

### Method 2: Manual Clone

```bash
cd ~/.claude
git clone https://github.com/hanzoai/agents.git
```

## 3. Verify Setup (1 minute)

In Claude Code or Claude Desktop:

```
You: "List available hanzo agents"
Claude: [Shows installed agents]

You: "Use dev agent to create a hello world function"
Claude: [dev agent creates simple, clean implementation]
```

## 4. Your First Agent Workflow (1 minute)

### Example: Build a Simple API

```
You: "Use full-stack-engineer to build a REST API for a todo app"

Claude (as full-stack-engineer):
I'll build a complete todo API with:
- Backend: FastAPI with PostgreSQL
- Database: Todo table with migrations
- Endpoints: CRUD operations
- Tests: Pytest with 90%+ coverage
- Deployment: Docker container

[Proceeds to implement with test-driven approach]
```

### Example: Fix a Bug

```
You: "Use senior-engineer to debug why users can't log in"

Claude (as senior-engineer):
Let me investigate:
1. Check authentication service logs
2. Trace login flow from API to database
3. Identify the failure point
4. Propose and implement fix
5. Add test to prevent regression

[Shows actual debugging process with logs and fix]
```

### Example: Review Code

```
You: "Use reviewer to check this pull request for issues"

Claude (as reviewer):
Code review findings:
✅ Good: Test coverage at 95%
⚠️  Security: SQL injection risk in line 42
⚠️  Performance: N+1 query in user fetch
❌ Critical: Missing authentication check

[Provides specific fixes for each issue]
```

## Core Team Agents

Start with these **7 essential agents** (covers 90% of development tasks):

| Agent | Role | Use For |
|-------|------|---------|
| **cto** | Technical Leadership | Architecture decisions, technology evaluation |
| **dev** | Primary Developer | Clean, simple implementations |
| **hanzo** | Infrastructure Specialist | Hanzo platform integration |
| **reviewer** | Code Quality | PR reviews, best practices |
| **architect** | System Design | Architecture planning, system design |
| **senior-engineer** | Advanced Development | Complex refactoring, optimization |
| **devops-troubleshooter** | Operations | Production debugging, deployment |

## Common Workflows

### Workflow 1: New Feature Development

```
You: "Implement user profile page with avatar upload"

Process:
1. architect → Design component structure
2. dev → Implement backend API
3. dev → Implement frontend UI
4. reviewer → Code review
5. devops-troubleshooter → Deploy to staging

Result: Complete, tested, deployed feature
```

### Workflow 2: Performance Optimization

```
You: "Our dashboard is loading slowly, optimize it"

Process:
1. senior-engineer → Profile and identify bottlenecks
2. dev → Implement caching layer
3. dev → Optimize database queries
4. reviewer → Validate changes
5. senior-engineer → Benchmark improvements

Result: Dashboard load time: 3.2s → 0.8s ✅
```

### Workflow 3: Security Audit

```
You: "Audit our authentication system for vulnerabilities"

Process:
1. security-auditor → Scan for OWASP top 10
2. reviewer → Review auth code
3. dev → Fix identified issues
4. security-auditor → Verify fixes

Result: All critical vulnerabilities patched ✅
```

## Role-Based Agents (Advanced)

For larger teams, use role-based agents that coordinate specialists:

```bash
# Install role-based agents
/plugin install hanzo-role-based
```

**Available roles:**
- `full-stack-engineer` - Complete application development
- `senior-engineer` - Complex technical challenges
- `tech-lead` - Team coordination and planning
- `staff-engineer` - Platform-level architecture

**Example usage:**
```
You: "Use tech-lead to plan our Q1 microservices migration"

Claude (as tech-lead):
[Creates comprehensive RFC with:]
- Phase 1: Service boundary identification (2 weeks)
- Phase 2: Extract user service (3 weeks)
- Phase 3: Extract payment service (3 weeks)
- Dependencies, risks, rollback plans
- Team assignments and timeline
```

## Agent Selection Guide

**When to use which agent:**

| Task | Agent | Why |
|------|-------|-----|
| Simple CRUD API | `dev` | Minimal, clean implementation |
| Complex refactoring | `senior-engineer` | Deep expertise needed |
| Multi-service feature | `full-stack-engineer` | Coordinates full stack |
| Sprint planning | `tech-lead` | Organizational skills |
| Platform design | `staff-engineer` | Company-wide impact |
| Code review | `reviewer` | Quality focus |
| Production bug | `devops-troubleshooter` | Incident response |
| Security audit | `security-auditor` (via reviewer) | Security expertise |

## Best Practices

### 1. Let Claude Select Agents

```
✅ Good: "Build a user authentication system"
(Claude automatically selects full-stack-engineer)

❌ Overthinking: "Use backend-architect, then frontend-developer, then..."
(Just describe what you want, let Claude coordinate)
```

### 2. Provide Context

```
✅ Good: "Optimize our Node.js API that handles 10k req/min, currently timing out at 5s"
(Clear context helps agent focus)

❌ Vague: "Make it faster"
(Unclear what to optimize)
```

### 3. Trust the Tests

```
✅ Good: "Show me the test results"
(Verify with actual output)

❌ Risky: "Looks good to me"
(Always verify)
```

### 4. Use Hanzo Tools First

```
✅ Good: "Use @hanzo/ui for the dashboard components"
(Leverages Hanzo infrastructure)

❌ Reinventing: "Build custom UI components from scratch"
(Wastes time)
```

## Next Steps

1. **Try the Core Team**: Start with `dev`, `reviewer`, `hanzo`
2. **Read Agent Docs**: Check `/Users/z/work/hanzo/agents/agents/README.md`
3. **Explore Workflows**: See `/Users/z/work/hanzo/agents/workflows/`
4. **Customize**: Create your own agents in `~/.claude/agents/`
5. **Share**: Contribute useful agents back to the community

## Example Session

```
$ claude

You: "I need to build a blog API with authentication"

Claude: I'll help you build that. Let me use the full-stack-engineer
agent to implement a complete blog API.

[Creates todo list:]
- Design database schema (posts, users, auth)
- Implement FastAPI backend with JWT auth
- Create CRUD endpoints for posts
- Add comprehensive tests
- Set up Docker deployment

[Proceeds to implement each task with tests passing]

Files Created:
- backend/api/main.py (FastAPI app)
- backend/api/auth.py (JWT authentication)
- backend/api/posts.py (Blog endpoints)
- backend/models.py (SQLAlchemy models)
- tests/test_api.py (95% coverage)
- docker-compose.yml (Deployment)

Tests Results:
=================== 24 passed in 2.3s ===================

You: "Great! Now use reviewer to check the code quality"

Claude (as reviewer):
Code Review - Blog API
✅ Security: JWT properly implemented
✅ Tests: Excellent coverage (95%)
⚠️  Performance: Consider adding database indexes
✅ Documentation: Clear API docs
✅ Error Handling: Comprehensive

Recommendation: Ready to deploy. Add indexes before production load.

You: "Perfect, deploy to Hanzo Cloud"

Claude: Using hanzo agent for deployment...
[Deploys to Hanzo Cloud with monitoring configured]

Deployed!
- URL: https://your-app.hanzo.cloud
- Monitoring: https://console.hanzo.ai/apps/your-app
- Logs: hanzo logs tail your-app

You're all set! 🚀
```

---

Start building with AI assistance today. The agents handle the complexity while you focus on what to build.
