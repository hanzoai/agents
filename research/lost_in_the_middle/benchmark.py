import numpy as np
import json

class LostInTheMiddleBenchmark:
    """
    Reproducible benchmark for HIP-1347.
    Simulates document retrieval and placement performance.
    """
    def __init__(self, num_docs=10):
        self.num_docs = num_docs

    def simulate_llm_scoring(self, relevant_pos):
        """
        Simulates the U-shaped performance curve observed in TACL 2024.
        """
        x = relevant_pos / (self.num_docs - 1)
        # Quadratic U-shape function: f(x) = 4(x-0.5)^2 + C
        # Yields high values at 0 and 1, low at 0.5
        performance = 0.8 * (4 * (x - 0.5)**2 + 0.2)
        return max(0, min(1, performance))

    def run_experiment(self, strategy="linear"):
        results = []
        for target_idx in range(self.num_docs):
            if strategy == "linear":
                # Target is placed at target_idx
                score = self.simulate_llm_scoring(target_idx)
            elif strategy == "u-shape":
                # Re-order logic: [1, 3, 5, ..., 4, 2]
                reordered_indices = self.get_u_shape_indices(self.num_docs)
                # Find where the 'most relevant' (index 0) ended up
                actual_pos = reordered_indices.index(target_idx)
                score = self.simulate_llm_scoring(actual_pos)
            results.append(score)
        return np.mean(results), results

    @staticmethod
    def get_u_shape_indices(n):
        """
        Creates a U-shape ordering of indices.
        Example for n=5: [0, 2, 4, 3, 1]
        """
        indices = list(range(n))
        reordered = []
        left = 0
        right = n - 1
        while left <= right:
            reordered.append(indices[left])
            left += 1
            if left <= right:
                reordered.insert(len(reordered), indices[left]) # This is a placeholder for actual logic
                # Real U-shape:
                # 1st best -> Start
                # 2nd best -> End
                # 3rd best -> Start + 1...
        
        # Correct U-Shape Logic
        final_order = [None] * n
        for i, original_idx in enumerate(range(n)):
            if i % 2 == 0:
                final_order[i // 2] = original_idx
            else:
                final_order[n - 1 - (i // 2)] = original_idx
        return final_order

if __name__ == "__main__":
    bench = LostInTheMiddleBenchmark(num_docs=20)
    
    lin_avg, lin_raw = bench.run_experiment("linear")
    u_avg, u_raw = bench.run_experiment("u-shape")
    
    print(f"Linear Strategy Avg Accuracy: {lin_avg:.4f}")
    print(f"U-Shape Strategy Avg Accuracy: {u_avg:.4f}")
    print(f"Improvement: {((u_avg/lin_avg)-1)*100:.2f}%")
    
    with open("results.json", "w") as f:
        json.dump({"linear": lin_raw, "u_shape": u_raw}, f)
