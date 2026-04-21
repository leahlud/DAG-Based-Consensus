# DAG-Based Consensus Simulation

A simplified implementation of a DAG-based ordering protocol with simulated validators that build a DAG of proposals and derive a total order.

<ins>Authors</ins>: Nithin Parthasarathy and Leah Ludwikowski

## Overview

Validators propose blocks simultaneously each round, referencing certified blocks from the previous round. Once a block collects `2f+1` signatures it becomes a certificate and enters the DAG. A sequencer then scans the DAG round by round and outputs a deterministic total ordering of certified blocks.

## Architecture

- **Network**: communication layer routing messages between validators
- **Validator**: proposes blocks, broadcasts to network, collects votes, certifies blocks at `2f+1` signatures. Each validator runs as a goroutine
- **DAG**: shared data structure storing certified blocks organized by round, protected by a mutex for concurrent access
- **Sequencer**: scans the DAG round by round and outputs a deterministic total ordering once a round has `2f+1` certified blocks

## Fault Model

- Tolerates up to `f` byzantine validators where `n ≥ 3f+1`
- Faulty set is re-randomized each round to simulate varying fault patterns
- Byzantine behavior is modeled as skipped proposals rather than active attacks

## Requirements

**Simulation**
- Go 1.25+

**Visualization**
- Python 3.13+
- `matplotlib`
- `networkx`
- `pandas`
- `numpy`

## Usage

### Run the simulation

```bash
go run main.go -f <int> -n <int> -rounds <int> -delay <int> -p <float>
```

This will output three CSV files:
- `edges.csv` — DAG edges (source, target, round, author)
- `order.csv` — total ordering of certified blocks
- `byzantine.csv` — byzantine status of each validator per round

### Visualize the results

```bash
python main.py
```

<ins>Output</ins>: `dag_visualization.png`

## Limitations

- No real cryptography: blocks are certified by simulation logic rather than actual signing and verification
- No mempool: block payloads are opaque and no transaction execution is modeled
- Simulated network: not an actual distributed network
