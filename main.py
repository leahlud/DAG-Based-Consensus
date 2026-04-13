import argparse
import numpy as np
import pandas as pd
import networkx as nx
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches

COLORS = {
    'committed': "#309FA7",
    'committedBorder': "#0F586E",
    'pending': '#B4B2A9',
    'pendingBorder': '#5F5E5A',
    'shading': '#E1F5EE',
    'arrows': "#9FDFE1",
    'byzantine': '#7D7D7D',
    'honest': "#2ECC71",
    'honestBorder': "#1C8B4A"
}

# load graph data from CSVs 
edges_df = pd.read_csv("edges.csv")
order_df = pd.read_csv("order.csv")
byz_df = pd.read_csv("byzantine.csv")

# parse user args
parser = argparse.ArgumentParser(description="Visualize DAG-based consensus.")
parser.add_argument("--f", type=int, default=1, help="Number of faulty validators f (default: 1)")
args = parser.parse_args()

F = args.f
QUORUM = 2 * F + 1  # 2f+1
 
# parse node names (i.e. "r2-v0" -> round=2, validator=0)
def parse_node(name):
    r, v = name.split("-")
    return int(r[1:]), int(v[1:])
 
# collect nodes and dimensions
all_nodes  = sorted(set(pd.concat([edges_df["source"], edges_df["target"]]).unique()), key=parse_node)
rounds     = sorted(set(parse_node(n)[0] for n in all_nodes))
validators = byz_df['validator_id'].unique()
n_vals     = len(validators)
 
# total ordering from CSV
order_map = dict(zip(order_df["block_id"], order_df["position"]))
 
# grid positions: x = round, y = validator (v0 at top), y_offset = where the validator info is located
X_STEP = 3
Y_STEP = 1
VALIDATOR_Y_OFFSET = 0.6
NODE_R = 0.38
 
def grid_pos(node):
    r, v = parse_node(node)
    return ((r - 1) * X_STEP, -v * Y_STEP)
 
pos = {n: grid_pos(n) for n in all_nodes}
 
# build graph
G = nx.DiGraph()
for _, row in edges_df.iterrows():
    G.add_edge(row["source"], row["target"])
 
# plot
fig, (ax, ax2) = plt.subplots(
    2, 1,
    figsize=(min(1 + len(rounds) * 3, 15), min(1 + n_vals * 1.5, 8.5)),
    gridspec_kw={'height_ratios': [4.5, 1]},
    sharex=True
)
fig.patch.set_facecolor("#f8f8f6")
ax.set_facecolor("#f8f8f6")
 
# round shading
for r in rounds:
    x = (r - 1) * X_STEP
    ax.axvspan(x - 1.2, x + 1.2, color=COLORS['shading'], zorder=0, alpha=0.8)

# edges
for src, tgt in G.edges():
    x1, y1 = pos[src]
    x2, y2 = pos[tgt]
    dx, dy = x2 - x1, y2 - y1
    dist = np.hypot(dx, dy)
    if dist == 0:
        continue
    ux, uy = dx / dist, dy / dist
    sx, sy = x1 + ux * NODE_R, y1 + uy * NODE_R
    ex, ey = x2 - ux * NODE_R, y2 - uy * NODE_R
    ax.annotate("", xy=(ex, ey), xytext=(sx, sy),
                arrowprops=dict(
                    arrowstyle="-|>",
                    color=COLORS['arrows'],
                    lw=1.2,
                    connectionstyle="arc3,rad=0.0",
                    shrinkA=0, shrinkB=0,
                ))
 
# nodes
NODE_R = 0.38
for node in all_nodes:
    x, y = pos[node]
    r, v  = parse_node(node)
    committed = True
    color  = COLORS['committed'] if committed else COLORS['pending']
    border = COLORS['committedBorder'] if committed else COLORS['pendingBorder']
    ax.add_patch(plt.Circle((x, y), NODE_R, color=color, ec=border, lw=1.5, zorder=3))
    ax.text(x, y, node, ha="center", va="center", fontsize=6.5, fontweight="bold", color="white", zorder=4)

 
# TOP PLOT 
# round labels (top)
for r in rounds:
    x      = (r - 1) * X_STEP
    ax.text(x, 0.9, f"Round {r}", ha="center", va="bottom", fontsize=10, color="#333", fontweight="500")
 
# validator labels (left)
for v in validators:
    ax.text(-1.1, -v * Y_STEP, f"V{v}", ha="center", va="center", fontsize=9, color="#555")
 
ax.set_xlim(-1.5, (max(rounds) - 1) * X_STEP + 1.5)
ax.set_ylim(-n_vals * Y_STEP - 0.2, 1.3)
ax.set_aspect("equal")
ax.axis("off")
ax.set_title("DAG Consensus Results", fontsize=13, pad=12)
 
# BOTTOM PLOT
# validators
ax2.set_facecolor("#f8f8f6")
CELL_W = X_STEP
CELL_H = 0.8

for r in rounds:
    for v in validators:
        x = (r - 1) * X_STEP
        y = -v
        
        rows = byz_df[
            (byz_df['round'] == r) & (byz_df['validator_id'] == v)
        ]

        is_byzantine = False
        if not rows.empty:
            is_byzantine = str(rows.iloc[0]['byzantine']).lower() == "true"

        color = COLORS['byzantine'] if is_byzantine else COLORS['honest']
        border = COLORS['pendingBorder'] if is_byzantine else COLORS['honestBorder']

        rect = plt.Rectangle(
            (x - 0.5, y - 0.4),
            1.0,
            CELL_H,
            color=color,
            ec=border,
            lw=1
        )
        ax2.add_patch(rect)
        ax2.text(x, y, f"V{v}", ha="center", va="center", fontsize=7, color="white")

ax2.set_xlim(-1.5, (max(rounds) - 1) * X_STEP + 1.5)
ax2.set_ylim(-len(validators), 1)
ax2.set_xticks([(r - 1) * X_STEP for r in rounds])
ax2.set_xticklabels([f"R{r}" for r in rounds], fontsize=8)
ax2.set_yticks([-v for v in validators])
ax2.set_yticklabels([f"V{v}" for v in validators], fontsize=8)

ax2.axis("off")
ax2.set_title("Validator Byzantine Status", fontsize=10)

# legends
ax.legend(handles=[
    mpatches.Patch(color=COLORS['committed'], label="Committed block"),
    mpatches.Patch(color=COLORS['honest'], label="Honest"),
    mpatches.Patch(color=COLORS['byzantine'], label="Byzantine"),
], loc="lower right", fontsize=8, framealpha=0.9)
 
# final plot
plt.tight_layout()
plt.savefig("dag_visualization.png", dpi=150, bbox_inches="tight")
plt.show()