#!/usr/bin/env python3
"""Generate plugin entries for all specialist agents."""

import json
import re
from pathlib import Path
from typing import Dict, List, Optional


def extract_frontmatter(content: str) -> Optional[Dict[str, str]]:
    """Extract YAML frontmatter from markdown file."""
    match = re.match(r'^---\s*\n(.*?)\n---\s*\n', content, re.DOTALL)
    if not match:
        return None

    frontmatter = {}
    for line in match.group(1).strip().split('\n'):
        if ':' in line:
            key, value = line.split(':', 1)
            frontmatter[key.strip()] = value.strip()

    return frontmatter


def generate_keywords(agent_name: str, description: str) -> List[str]:
    """Generate relevant keywords based on agent name and description."""
    # Language-specific agents
    language_keywords = {
        'python-pro': ['python', 'programming', 'async', 'fastapi'],
        'javascript-pro': ['javascript', 'programming', 'nodejs', 'frontend'],
        'typescript-pro': ['typescript', 'programming', 'types', 'nodejs'],
        'golang-pro': ['go', 'golang', 'programming', 'backend'],
        'rust-pro': ['rust', 'programming', 'systems', 'performance'],
        'java-pro': ['java', 'programming', 'spring', 'backend'],
        'c-pro': ['c', 'programming', 'systems', 'embedded'],
        'cpp-pro': ['cpp', 'c++', 'programming', 'systems'],
        'csharp-pro': ['csharp', 'c#', 'dotnet', 'programming'],
        'ruby-pro': ['ruby', 'programming', 'rails', 'backend'],
        'php-pro': ['php', 'programming', 'web', 'backend'],
        'scala-pro': ['scala', 'programming', 'jvm', 'functional'],
        'elixir-pro': ['elixir', 'programming', 'functional', 'erlang'],
        'sql-pro': ['sql', 'database', 'queries', 'optimization'],
    }

    # Framework-specific agents
    framework_keywords = {
        'fastapi-pro': ['fastapi', 'python', 'api', 'async'],
        'django-pro': ['django', 'python', 'web', 'backend'],
        'flutter-expert': ['flutter', 'dart', 'mobile', 'crossplatform'],
        'unity-developer': ['unity', 'gamedev', 'csharp', '3d'],
        'minecraft-bukkit-pro': ['minecraft', 'bukkit', 'java', 'plugins'],
    }

    # Role-specific agents
    role_keywords = {
        'backend-architect': ['backend', 'api', 'architecture', 'microservices'],
        'frontend-developer': ['frontend', 'ui', 'react', 'web'],
        'mobile-developer': ['mobile', 'ios', 'android', 'apps'],
        'ios-developer': ['ios', 'swift', 'mobile', 'apple'],
        'ai-engineer': ['ai', 'llm', 'ml', 'rag', 'agents'],
        'ml-engineer': ['ml', 'machine-learning', 'ai', 'training'],
        'mlops-engineer': ['mlops', 'ml', 'deployment', 'monitoring'],
        'data-engineer': ['data', 'etl', 'pipeline', 'spark'],
        'data-scientist': ['data-science', 'ml', 'analytics', 'statistics'],
        'database-architect': ['database', 'schema', 'design', 'optimization'],
        'database-admin': ['database', 'administration', 'performance', 'backup'],
        'database-optimizer': ['database', 'optimization', 'performance', 'tuning'],
        'cloud-architect': ['cloud', 'aws', 'azure', 'gcp', 'infrastructure'],
        'hybrid-cloud-architect': ['cloud', 'hybrid', 'multi-cloud', 'infrastructure'],
        'kubernetes-architect': ['kubernetes', 'k8s', 'containers', 'orchestration'],
        'terraform-specialist': ['terraform', 'iac', 'infrastructure', 'cloud'],
        'devops-troubleshooter': ['devops', 'troubleshooting', 'debugging', 'operations'],
        'deployment-engineer': ['deployment', 'ci-cd', 'release', 'automation'],
        'observability-engineer': ['observability', 'monitoring', 'logging', 'metrics'],
        'performance-engineer': ['performance', 'optimization', 'scalability', 'profiling'],
        'network-engineer': ['network', 'infrastructure', 'security', 'protocols'],
        'security-auditor': ['security', 'audit', 'vulnerabilities', 'compliance'],
        'backend-security-coder': ['security', 'backend', 'api', 'encryption'],
        'frontend-security-coder': ['security', 'frontend', 'xss', 'csrf'],
        'mobile-security-coder': ['security', 'mobile', 'ios', 'android'],
        'incident-responder': ['incident', 'security', 'response', 'forensics'],
        'blockchain-developer': ['blockchain', 'web3', 'smart-contracts', 'crypto'],
        'graphql-architect': ['graphql', 'api', 'schema', 'queries'],
        'payment-integration': ['payment', 'stripe', 'fintech', 'transactions'],
        'code-reviewer': ['code-review', 'quality', 'best-practices', 'feedback'],
        'debugger': ['debugging', 'troubleshooting', 'errors', 'fixes'],
        'error-detective': ['errors', 'debugging', 'root-cause', 'investigation'],
        'test-automator': ['testing', 'automation', 'qa', 'ci-cd'],
        'tdd-orchestrator': ['tdd', 'testing', 'test-driven', 'quality'],
        'legacy-modernizer': ['legacy', 'modernization', 'refactoring', 'migration'],
        'api-documenter': ['documentation', 'api', 'openapi', 'developer-experience'],
        'docs-architect': ['documentation', 'architecture', 'technical-writing', 'knowledge'],
        'tutorial-engineer': ['tutorials', 'education', 'documentation', 'learning'],
        'ui-ux-designer': ['ui', 'ux', 'design', 'user-experience'],
        'ui-visual-validator': ['ui', 'visual', 'testing', 'validation'],
        'dx-optimizer': ['developer-experience', 'dx', 'tooling', 'productivity'],
        'architect': ['architecture', 'system-design', 'scalability', 'patterns'],
        'architect-review': ['architecture', 'review', 'design', 'patterns'],
        'business-analyst': ['business', 'analysis', 'requirements', 'stakeholders'],
        'quant-analyst': ['quantitative', 'finance', 'trading', 'analytics'],
        'risk-manager': ['risk', 'compliance', 'security', 'governance'],
        'content-marketer': ['content', 'marketing', 'seo', 'copywriting'],
        'customer-support': ['support', 'customer-service', 'help', 'documentation'],
        'hr-pro': ['hr', 'human-resources', 'recruiting', 'talent'],
        'legal-advisor': ['legal', 'compliance', 'contracts', 'privacy'],
        'sales-automator': ['sales', 'automation', 'crm', 'outreach'],
        'context-manager': ['context', 'memory', 'state', 'management'],
        'mermaid-expert': ['mermaid', 'diagrams', 'visualization', 'documentation'],
        'prompt-engineer': ['prompts', 'ai', 'llm', 'optimization'],
        'reference-builder': ['references', 'citations', 'documentation', 'knowledge'],
        'reviewer': ['review', 'feedback', 'quality', 'assessment'],
        'scientist': ['research', 'science', 'analysis', 'methodology'],
        'search-specialist': ['search', 'information-retrieval', 'indexing', 'ranking'],
    }

    # SEO-specific agents
    seo_keywords = {
        'seo-authority-builder': ['seo', 'authority', 'backlinks', 'content'],
        'seo-cannibalization-detector': ['seo', 'cannibalization', 'content', 'analysis'],
        'seo-content-auditor': ['seo', 'audit', 'content', 'quality'],
        'seo-content-planner': ['seo', 'planning', 'content', 'strategy'],
        'seo-content-refresher': ['seo', 'refresh', 'content', 'updates'],
        'seo-content-writer': ['seo', 'writing', 'content', 'copywriting'],
        'seo-keyword-strategist': ['seo', 'keywords', 'research', 'strategy'],
        'seo-meta-optimizer': ['seo', 'metadata', 'optimization', 'titles'],
        'seo-snippet-hunter': ['seo', 'snippets', 'serp', 'ranking'],
        'seo-structure-architect': ['seo', 'structure', 'architecture', 'navigation'],
    }

    # Core agents
    core_keywords = {
        'bot': ['automation', 'cli', 'tools', 'scripting'],
        'dev': ['development', 'coding', 'programming', 'tools'],
        'hanzo': ['hanzo', 'platform', 'ai', 'infrastructure'],
        'neo': ['assistant', 'ai', 'helper', 'productivity'],
        'vp': ['leadership', 'management', 'strategy', 'oversight'],
        'cto': ['cto', 'leadership', 'technology', 'strategy'],
    }

    # Combine all keyword dictionaries
    all_keywords = {
        **language_keywords,
        **framework_keywords,
        **role_keywords,
        **seo_keywords,
        **core_keywords
    }

    # Get specific keywords or generate generic ones
    if agent_name in all_keywords:
        return all_keywords[agent_name]

    # Generic keywords based on description analysis
    keywords = ['agent', 'ai']

    # Add keywords from description
    desc_lower = description.lower()
    if 'api' in desc_lower:
        keywords.append('api')
    if 'test' in desc_lower:
        keywords.append('testing')
    if 'security' in desc_lower:
        keywords.append('security')
    if 'performance' in desc_lower:
        keywords.append('performance')
    if 'cloud' in desc_lower:
        keywords.append('cloud')
    if 'database' in desc_lower or 'db' in desc_lower:
        keywords.append('database')

    return keywords[:4]  # Limit to 4 keywords


def generate_plugin_entry(agent_file: Path) -> Optional[Dict]:
    """Generate a plugin entry for an agent file."""
    content = agent_file.read_text()
    frontmatter = extract_frontmatter(content)

    if not frontmatter or 'name' not in frontmatter or 'description' not in frontmatter:
        return None

    agent_name = frontmatter['name']
    description = frontmatter['description']
    keywords = generate_keywords(agent_name, description)

    plugin = {
        "name": f"hanzo-{agent_name}",
        "source": "./",
        "description": description,
        "version": "1.0.1",
        "author": {
            "name": "Hanzo AI",
            "email": "dev@hanzo.ai",
            "url": "https://github.com/hanzoai"
        },
        "homepage": "https://github.com/hanzoai/agents",
        "repository": "https://github.com/hanzoai/agents",
        "license": "MIT",
        "keywords": keywords,
        "category": "specialists",
        "strict": False,
        "agents": [f"./agents/{agent_file.name}"]
    }

    return plugin


def main():
    """Generate plugin entries for all agent files."""
    agents_dir = Path(__file__).parent / 'agents'

    # Get all markdown files except README
    agent_files = sorted([
        f for f in agents_dir.glob('*.md')
        if f.name != 'README.md'
    ])

    plugins = []
    for agent_file in agent_files:
        plugin = generate_plugin_entry(agent_file)
        if plugin:
            plugins.append(plugin)
            print(f"✓ Generated plugin for {plugin['name']}")
        else:
            print(f"✗ Skipped {agent_file.name} (no valid frontmatter)")

    # Write to JSON file
    output_file = Path(__file__).parent / 'specialist-plugins.json'
    with output_file.open('w') as f:
        json.dump(plugins, f, indent=2)

    print(f"\n✓ Generated {len(plugins)} plugin entries")
    print(f"✓ Output written to {output_file}")


if __name__ == '__main__':
    main()
