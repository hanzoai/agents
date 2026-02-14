/**
 * Attachment Type Definitions
 *
 * Re-exports all attachment types from @hanzo/agents-shared.
 * This file is kept for backwards compatibility with existing imports.
 */

export type {
  BaseAttachment,
  LinearIssueAttachment,
  TerminalAttachment,
  WorkspaceMetadataAttachment,
} from '@hanzo/agents-shared';

export {
  createLinearIssueAttachment,
  createWorkspaceMetadataAttachment,
  isLinearIssueAttachment,
  isWorkspaceMetadataAttachment,
} from '@hanzo/agents-shared';
