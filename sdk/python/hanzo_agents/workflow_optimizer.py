"""
Automatic Prompt & Multi-Call Workflow Optimizer (DSPy Clean-Room Teleprompter)
Implementation for Hanzo AI Bounty #37 / HIP-1347: Event-Driven Agent Reliability.

Optimizes agent instructions and multi-call programs against executable metrics.
Measures held-out pass rate, dollar cost, token count, and latency with HIP-1334 run records.
"""

import time
import json
import uuid
import hashlib
from typing import Dict, List, Any, Callable, Optional, Tuple
from dataclasses import dataclass, asdict, field


@dataclass
class Example:
    inputs: Dict[str, Any]
    expected_output: Any
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class CandidateMetrics:
    pass_rate: float
    avg_latency_ms: float
    total_tokens: int
    cost_usd: float
    eval_count: int


@dataclass
class OptimizationRunRecord:
    run_id: str
    target_program: str
    baseline_metrics: CandidateMetrics
    optimized_metrics: CandidateMetrics
    delta_pass_rate: float
    delta_cost_usd: float
    delta_latency_ms: float
    timestamp: str
    pareto_optimal: bool

    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)


class PromptOptimizer:
    """
    Automated Teleprompter & Workflow Optimizer.
    Evaluates candidate instructions against task metrics and synthesizes optimized prompt variants.
    """

    def __init__(
        self,
        cost_per_1k_tokens_usd: float = 0.0015,
        target_metric: Optional[Callable[[Any, Any], float]] = None
    ):
        self.cost_per_1k_tokens = cost_per_1k_tokens_usd
        self.metric_fn = target_metric or self._default_exact_match

    @staticmethod
    def _default_exact_match(predicted: Any, expected: Any) -> float:
        if str(predicted).strip().lower() == str(expected).strip().lower():
            return 1.0
        return 0.0

    def evaluate_candidate(
        self,
        instruction_prompt: str,
        program_runner: Callable[[str, Dict[str, Any]], Tuple[Any, int, float]],
        validation_set: List[Example]
    ) -> CandidateMetrics:
        """
        Runs the candidate instruction across validation examples.
        program_runner must return: (output, tokens_used, latency_ms)
        """
        if not validation_set:
            raise ValueError("Validation set cannot be empty")

        total_score = 0.0
        total_tokens = 0
        total_latency = 0.0

        for ex in validation_set:
            start_t = time.time()
            output, tokens, lat = program_runner(instruction_prompt, ex.inputs)
            dur_ms = lat if lat > 0 else (time.time() - start_t) * 1000.0

            score = self.metric_fn(output, ex.expected_output)
            total_score += score
            total_tokens += tokens
            total_latency += dur_ms

        n = len(validation_set)
        pass_rate = round(total_score / n, 4)
        avg_lat = round(total_latency / n, 2)
        total_cost = round((total_tokens / 1000.0) * self.cost_per_1k_tokens, 6)

        return CandidateMetrics(
            pass_rate=pass_rate,
            avg_latency_ms=avg_lat,
            total_tokens=total_tokens,
            cost_usd=total_cost,
            eval_count=n
        )

    def bootstrap_optimized_prompt(
        self,
        base_instruction: str,
        training_pool: List[Example],
        validation_set: List[Example],
        program_runner: Callable[[str, Dict[str, Any]], Tuple[Any, int, float]],
        max_demonstrations: int = 3
    ) -> Tuple[str, OptimizationRunRecord]:
        """
        Generates few-shot demonstrations from validated successes (DSPy BootstrapFewShot pattern),
        evaluates baseline vs optimized, and produces HIP-1334 compliant run records.
        """
        # 1. Baseline Evaluation
        baseline_metrics = self.evaluate_candidate(base_instruction, program_runner, validation_set)

        # 2. Extract effective demonstrations from training pool
        demonstrations = []
        for ex in training_pool:
            out, _, _ = program_runner(base_instruction, ex.inputs)
            if self.metric_fn(out, ex.expected_output) == 1.0:
                demonstrations.append((ex.inputs, out))
                if len(demonstrations) >= max_demonstrations:
                    break

        # 3. Synthesize candidate prompt with structured few-shot context
        demo_blocks = []
        for i, (inp, out) in enumerate(demonstrations, 1):
            demo_blocks.append(f"Example {i}:\nInput: {json.dumps(inp)}\nOutput: {out}")
        
        demo_text = "\n\n".join(demo_blocks)
        optimized_prompt = f"{base_instruction.strip()}\n\nVerified Demonstrations:\n{demo_text}\n"

        # 4. Evaluated Optimized Candidate
        optimized_metrics = self.evaluate_candidate(optimized_prompt, program_runner, validation_set)

        # 5. Build HIP-1334 Research Run Record
        delta_pass = round(optimized_metrics.pass_rate - baseline_metrics.pass_rate, 4)
        delta_cost = round(optimized_metrics.cost_usd - baseline_metrics.cost_usd, 6)
        delta_lat = round(optimized_metrics.avg_latency_ms - baseline_metrics.avg_latency_ms, 2)
        pareto = (delta_pass >= 0.0)

        record = OptimizationRunRecord(
            run_id=f"opt_run_{uuid.uuid4().hex[:10]}",
            target_program="agent_instruction_optimizer",
            baseline_metrics=baseline_metrics,
            optimized_metrics=optimized_metrics,
            delta_pass_rate=delta_pass,
            delta_cost_usd=delta_cost,
            delta_latency_ms=delta_lat,
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            pareto_optimal=pareto
        )

        return optimized_prompt, record
