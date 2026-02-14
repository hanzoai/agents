/**
 * Conversation Type Definitions
 *
 * Re-exports conversation types from @hanzo/agents-shared.
 * This file is kept for backwards compatibility with existing imports.
 */

export type {
  AssistantMessageEntry,
  AssistantMessageGroup,
  // Message types
  ClaudeMessage,
  ConversationEntry,
  FileHistorySnapshotEntry,
  MessageContent,
  MessageGroup,
  // Entry types
  QueueOperationEntry,
  // Content types
  TextContent,
  ThinkingContent,
  ToolResultContent,
  ToolUseContent,
  UserMessageEntry,
  // Group types
  UserMessageGroup,
} from '@hanzo/agents-shared';

export {
  isAssistantMessageEntry,
  isFileHistorySnapshotEntry,
  isQueueOperationEntry,
  isTextContent,
  isThinkingContent,
  isToolResultContent,
  isToolUseContent,
  // Type guards
  isUserMessageEntry,
} from '@hanzo/agents-shared';
