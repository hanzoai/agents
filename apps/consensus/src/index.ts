export { AgentConsensus } from './consensus.js';
export type { AgentCallFn, AgentConfig, AgentVoteFn, Vote } from './consensus.js';

export { runMeshConsensus } from './mcp-bridge.js';
export type { McpClient } from './mcp-bridge.js';

export {
  computeConsensusScore,
  hasReachedFinality,
  synthesizeTopK,
} from './quasar-voting.js';
export type { AgentProposal, ConsensusResult, QuasarVote } from './quasar-voting.js';
