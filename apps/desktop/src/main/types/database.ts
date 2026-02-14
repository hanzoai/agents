/**
 * Database Type Definitions
 *
 * Re-exports canvas and node types from @hanzo/agents-shared.
 * This file is kept for backwards compatibility with existing imports.
 */

// Re-export AgentNodeData for convenience (also available from agent-node types)
// Re-export attachment types used in node data
export type {
  AgentNodeData,
  CanvasEdge,
  CanvasMetadata,
  CanvasNode,
  // Canvas types
  CanvasNodeType,
  CanvasState,
  CustomNodeData,
  NodeData,
  TerminalAttachment,
  // Node data types
  TerminalNodeData,
  Viewport,
} from '@hanzo/agents-shared';
export {
  // Type guards
  isAgentNodeData,
  isTerminalNodeData,
} from '@hanzo/agents-shared';
