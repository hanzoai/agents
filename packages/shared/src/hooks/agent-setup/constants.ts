/**
 * Constants for Agent Hooks Setup
 *
 * Defines environment variable names, ports, and markers used
 * for terminal-to-orchestrator communication.
 */

/**
 * Default port for the agent hooks HTTP server
 */
export const DEFAULT_HOOKS_PORT = 31415;

/**
 * Marker used to identify Hanzo Agents terminals
 */
export const TERMINAL_MARKER = 'HANZO_AGENTS';

/**
 * Environment variable names injected into terminal processes
 */
export const ENV_VARS = {
  /** Terminal ID for routing events back */
  TERMINAL_ID: 'HANZO_AGENTS_TERMINAL_ID',
  /** Workspace path for context */
  WORKSPACE_PATH: 'HANZO_AGENTS_WORKSPACE_PATH',
  /** Git branch for context */
  GIT_BRANCH: 'HANZO_AGENTS_GIT_BRANCH',
  /** Agent ID for tracking */
  AGENT_ID: 'HANZO_AGENTS_AGENT_ID',
  /** Port for HTTP callbacks */
  PORT: 'HANZO_AGENTS_PORT',
  /** Marker to detect orchestrator terminals */
  MARKER: 'HANZO_AGENTS_MARKER',
} as const;

/**
 * Directory names for hooks setup
 */
export const HOOKS_DIRS = {
  /** Base directory under home */
  BASE: '.hanzo-agents',
  /** Hooks scripts directory */
  HOOKS: 'hooks',
} as const;

/**
 * Script file names
 */
export const SCRIPT_FILES = {
  /** Notification script called by agent hooks */
  NOTIFY: 'notify.sh',
  /** Claude wrapper script */
  CLAUDE_WRAPPER: 'claude-wrapper.sh',
  /** Claude settings file */
  CLAUDE_SETTINGS: 'claude-settings.json',
} as const;
