# Specialist Agent Plugins

This repository contains 92+ specialist AI agents that can be installed as plugins for Claude Desktop and other AI platforms.

## Generated Plugins

The `specialist-plugins.json` file contains plugin definitions for all specialist agents in the `/agents` directory.

### Plugin Structure

Each plugin follows this structure:

```json
{
  "name": "hanzo-{agent-name}",
  "source": "./",
  "description": "{agent description}",
  "version": "1.0.1",
  "author": {
    "name": "Hanzo AI",
    "email": "dev@hanzo.ai",
    "url": "https://github.com/hanzoai"
  },
  "homepage": "https://github.com/hanzoai/agents",
  "repository": "https://github.com/hanzoai/agents",
  "license": "MIT",
  "keywords": ["{relevant-keywords}"],
  "category": "specialists",
  "strict": false,
  "agents": ["./agents/{filename}.md"]
}
```

## Available Categories

### Programming Languages (13 agents)
- **hanzo-python-pro** - Python 3.12+ with modern async, FastAPI, uv, ruff
- **hanzo-javascript-pro** - JavaScript/Node.js with modern ES features
- **hanzo-typescript-pro** - TypeScript with strict type safety
- **hanzo-golang-pro** - Go with goroutines and modern patterns
- **hanzo-rust-pro** - Rust with systems programming and performance
- **hanzo-java-pro** - Java with Spring Boot and modern JVM
- **hanzo-c-pro** - C with systems programming and embedded
- **hanzo-cpp-pro** - C++ with modern C++17/20 features
- **hanzo-csharp-pro** - C# with .NET Core and async patterns
- **hanzo-ruby-pro** - Ruby with Rails and modern patterns
- **hanzo-php-pro** - PHP with Laravel and modern features
- **hanzo-scala-pro** - Scala with functional programming
- **hanzo-elixir-pro** - Elixir with OTP and concurrency
- **hanzo-sql-pro** - SQL optimization and database queries

### Frameworks (4 agents)
- **hanzo-fastapi-pro** - FastAPI with async Python
- **hanzo-django-pro** - Django full-stack development
- **hanzo-flutter-expert** - Flutter cross-platform mobile
- **hanzo-unity-developer** - Unity game development

### Backend & Architecture (8 agents)
- **hanzo-backend-architect** - Scalable API design and microservices
- **hanzo-architect** - System design and architecture patterns
- **hanzo-architect-review** - Architecture review and validation
- **hanzo-graphql-architect** - GraphQL API design
- **hanzo-api-documenter** - OpenAPI and developer portals
- **hanzo-database-architect** - Database schema design
- **hanzo-database-admin** - Database administration
- **hanzo-database-optimizer** - Query and performance optimization

### AI & Machine Learning (5 agents)
- **hanzo-ai-engineer** - LLM apps, RAG, agents, multimodal AI
- **hanzo-ml-engineer** - Machine learning training and deployment
- **hanzo-mlops-engineer** - MLOps pipelines and monitoring
- **hanzo-data-engineer** - Data pipelines and ETL
- **hanzo-data-scientist** - Data analysis and modeling
- **hanzo-prompt-engineer** - AI prompt optimization

### Cloud & Infrastructure (9 agents)
- **hanzo-cloud-architect** - AWS, Azure, GCP architecture
- **hanzo-hybrid-cloud-architect** - Multi-cloud and hybrid strategies
- **hanzo-kubernetes-architect** - K8s orchestration and design
- **hanzo-terraform-specialist** - Infrastructure as Code
- **hanzo-deployment-engineer** - CI/CD and deployment automation
- **hanzo-devops-troubleshooter** - Operations debugging
- **hanzo-observability-engineer** - Monitoring, logging, tracing
- **hanzo-performance-engineer** - Performance optimization
- **hanzo-network-engineer** - Network infrastructure

### Frontend & Mobile (5 agents)
- **hanzo-frontend-developer** - React, Next.js, modern frontend
- **hanzo-mobile-developer** - iOS and Android development
- **hanzo-ios-developer** - Swift and iOS native development
- **hanzo-ui-ux-designer** - Design systems and user experience
- **hanzo-ui-visual-validator** - Visual testing and validation

### Security (5 agents)
- **hanzo-security-auditor** - Security audits and compliance
- **hanzo-backend-security-coder** - Backend security patterns
- **hanzo-frontend-security-coder** - XSS, CSRF, client security
- **hanzo-mobile-security-coder** - Mobile app security
- **hanzo-incident-responder** - Security incident response

### Testing & Quality (5 agents)
- **hanzo-code-reviewer** - Code review and best practices
- **hanzo-test-automator** - Test automation frameworks
- **hanzo-tdd-orchestrator** - Test-driven development
- **hanzo-debugger** - Debugging and troubleshooting
- **hanzo-error-detective** - Error investigation and root cause

### SEO & Content (10 agents)
- **hanzo-seo-keyword-strategist** - Keyword research and strategy
- **hanzo-seo-content-writer** - SEO-optimized content creation
- **hanzo-seo-content-planner** - Content calendar and planning
- **hanzo-seo-content-auditor** - Content quality audits
- **hanzo-seo-content-refresher** - Content updates and optimization
- **hanzo-seo-meta-optimizer** - Metadata and titles
- **hanzo-seo-structure-architect** - Site structure and navigation
- **hanzo-seo-authority-builder** - Backlinks and authority
- **hanzo-seo-snippet-hunter** - Featured snippets optimization
- **hanzo-seo-cannibalization-detector** - Content overlap detection

### Documentation (3 agents)
- **hanzo-docs-architect** - Documentation architecture
- **hanzo-tutorial-engineer** - Tutorial and guide creation
- **hanzo-reference-builder** - Reference documentation

### Business & Specialized (14 agents)
- **hanzo-business-analyst** - Requirements and stakeholder management
- **hanzo-blockchain-developer** - Smart contracts and Web3
- **hanzo-payment-integration** - Payment systems and fintech
- **hanzo-quant-analyst** - Quantitative finance and trading
- **hanzo-risk-manager** - Risk assessment and compliance
- **hanzo-content-marketer** - Content marketing and SEO
- **hanzo-customer-support** - Support documentation and processes
- **hanzo-hr-pro** - HR processes and recruiting
- **hanzo-legal-advisor** - Legal compliance and contracts
- **hanzo-sales-automator** - Sales automation and CRM
- **hanzo-legacy-modernizer** - Legacy code modernization
- **hanzo-context-manager** - Context and memory management
- **hanzo-dx-optimizer** - Developer experience optimization
- **hanzo-minecraft-bukkit-pro** - Minecraft plugin development

### Utilities (5 agents)
- **hanzo-mermaid-expert** - Diagram generation
- **hanzo-search-specialist** - Search and information retrieval
- **hanzo-reviewer** - General review and feedback
- **hanzo-scientist** - Research methodology
- **hanzo-reference-builder** - Citations and references

### Core Agents (5 agents)
- **hanzo-bot** - Automation and CLI tools
- **hanzo-dev** - General development tasks
- **hanzo-hanzo** - Hanzo platform integration
- **hanzo-neo** - AI assistant and productivity
- **hanzo-cto** - Technical leadership

## Usage

### Adding to marketplace.json

To add these plugins to your marketplace, append the contents of `specialist-plugins.json` to your `marketplace.json` plugins array:

```bash
# Merge with existing marketplace
jq '.plugins += input' marketplace.json specialist-plugins.json > marketplace-updated.json
```

### Installing Individual Plugins

Each plugin can be installed independently by referencing its agent file:

```json
{
  "mcpServers": {
    "hanzo-python-pro": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-agent"],
      "env": {
        "AGENT_FILE": "/path/to/agents/agents/python-pro.md"
      }
    }
  }
}
```

## Regenerating Plugins

To regenerate the plugin list after adding or modifying agents:

```bash
python generate_plugins.py
```

This will:
1. Scan all `.md` files in `/agents` directory (excluding README.md)
2. Extract frontmatter (name, description)
3. Generate appropriate keywords based on agent specialty
4. Create plugin entries in `specialist-plugins.json`

## Plugin Metadata

All plugins share these properties:
- **Version**: 1.0.1
- **Author**: Hanzo AI (dev@hanzo.ai)
- **License**: MIT
- **Repository**: https://github.com/hanzoai/agents
- **Category**: specialists
- **Strict mode**: false (allows flexible invocation)

## Keywords Strategy

Keywords are automatically generated based on agent specialization:

- **Language agents**: Language name + "programming" + specific features
- **Framework agents**: Framework + language + domain
- **Role agents**: Domain + responsibilities + technologies
- **SEO agents**: "seo" + specific focus area + "content"
- **Security agents**: "security" + domain + specific concerns

## Notes

- The `vp.md` agent was skipped because it lacks frontmatter
- All other 92 agents were successfully processed
- Plugins reference relative paths (`./agents/*.md`)
- Agent files must have YAML frontmatter with `name` and `description` fields

## License

MIT License - See LICENSE file for details
