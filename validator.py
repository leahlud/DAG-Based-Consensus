import hashlib
import time
import random


class Vertex:
    def __init__(self, round: int, validator_id: int, transactions: list, references: set):
        self.round = round
        self.validator_id = validator_id
        self.transactions = transactions
        self.references = references          # set of (round, vid) tuples
        self.vertex_id = (round, validator_id)
        self.timestamp = time.time()

    def digest(self):
        content = f"{self.round}:{self.validator_id}:{sorted(self.references)}"
        return hashlib.sha256(content.encode()).hexdigest()[:8]

    def __repr__(self):
        return f"V(r={self.round}, vid={self.validator_id}, refs={len(self.references)})"

# Class for validators
class Validator:
    def __init__(self, validator_id: int, f: int):
        self.validator_id = validator_id
        self.f = f
        self.n = 3*f + 1
        self.dag = {}             # (round, vid) -> Vertex
        self.committed = set()    # set of (round, vid) already committed
        self.total_order = []     # committed vertex ids in order

    # Propose a block for the round (called outside the class)
    def propose(self, current_round: int, known_vertices: dict) -> Vertex:
        refs = {
            vid for vid in known_vertices
            if vid[0] == current_round - 1
        }
        txs = list(range(self.validator_id * 100, self.validator_id * 101))
        vertex = Vertex(current_round, self.validator_id, txs, refs)
        self.dag[vertex.vertex_id] = vertex
        return vertex

    # Simulate validating a block if all its references are known
    def validate(self, vertex: Vertex, known_vertices: dict) -> bool:
        return all(ref in known_vertices for ref in vertex.references)

    # Bullshark: commit vertices from 2 rounds ago
    def try_commit(self, current_round: int, known_vertices: dict):
        commit_round = current_round - 2
        if commit_round < 0:
            return

        to_commit = sorted(
            [vid for vid in known_vertices
             if vid[0] == commit_round and vid not in self.committed],      # get uncommitted validators from the round we want to commit
            key=lambda vid: vid[1]    # sort by validator_id to make sure order isn't random
        )

        for vid in to_commit:
            self.committed.add(vid)
            self.total_order.append(vid)