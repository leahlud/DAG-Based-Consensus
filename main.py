import argparse
import numpy as np
import pandas as pd
import networkx as nx
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches

# --- constants ---
X_STEP = 3          # x = round
Y_STEP = 1          # y = validator (V0 at top)
NODE_R = 0.38

COLORS = {
    'committed':      "#309FA7",
    'committedBorder':"#0F586E",
    'shading':        '#E1F5EE',
    'arrows':         "#9FDFE1",
    'byzantine':      '#7D7D7D',
    'byzantineBorder':'#5F5E5A',
    'honest':         "#2ECC71",
    'honestBorder':   "#1C8B4A",
}

# --- args ---
parser = argparse.ArgumentParser(description="Visualize DAG-based consensus.")
parser.add_argument("--f", type=int, default=1, help="Number of faulty validators (default: 1)")
args = parser.parse_args()

F      = args.f
QUORUM = 2 * F + 1

# --- load data ---
edges_df = pd.read_csv("edges.csv")
order_df = pd.read_csv("order.csv")
byz_df   = pd.read_csv("byzantine.csv")

# --- parsing ---
def parse_block_name(name):
    r, v = name.split("-")
    return int(r[1:]), int(v[1:])

def grid_pos(block):
    r, v = parse_block_name(block)
    return ((r - 1) * X_STEP, -v * Y_STEP)

all_blocks = sorted(set(pd.concat([edges_df["source"], edges_df["target"]]).unique()), key=parse_block_name)
rounds     = sorted(set(parse_block_name(b)[0] for b in all_blocks))
validators = sorted(byz_df['validator_id'].unique())
n_vals     = len(validators)
pos        = {b: grid_pos(b) for b in all_blocks}

G = nx.DiGraph()
for _, row in edges_df.iterrows():
    G.add_edge(row["source"], row["target"])

# --- drawing functions ---
def draw_shading(ax):
    for r in rounds:
        ax.axvspan((r - 1) * X_STEP - 1.2, (r - 1) * X_STEP + 1.2,
                   color=COLORS['shading'], zorder=0, alpha=0.8)

def draw_edges(ax):
    for src, tgt in G.edges():
        x1, y1 = pos[src]
        x2, y2 = pos[tgt]
        dx, dy = x2 - x1, y2 - y1
        dist = np.hypot(dx, dy)
        if dist == 0:
            continue
        ux, uy = dx / dist, dy / dist
        ax.annotate("", xy=(x2 - ux * NODE_R, y2 - uy * NODE_R),
                        xytext=(x1 + ux * NODE_R, y1 + uy * NODE_R),
                    arrowprops=dict(arrowstyle="-|>", color=COLORS['arrows'],
                                    lw=1.2, connectionstyle="arc3,rad=0.0",
                                    shrinkA=0, shrinkB=0))

def draw_blocks(ax):
    for block in all_blocks:
        x, y = pos[block]
        ax.add_patch(plt.Circle((x, y), NODE_R, color=COLORS['committed'],
                                ec=COLORS['committedBorder'], lw=1.5, zorder=3))
        ax.text(x, y, block, ha="center", va="center",
                fontsize=6.5, fontweight="bold", color="white", zorder=4)

def draw_byzantine_grid(ax):
    for r in rounds:
        for v in validators:
            x = (r - 1) * X_STEP
            y = -validators.index(v)
            rows = byz_df[(byz_df['round'] == r) & (byz_df['validator_id'] == v)]
            is_byzantine = (not rows.empty) and str(rows.iloc[0]['byzantine']).lower() == "true"
            color  = COLORS['byzantine'] if is_byzantine else COLORS['honest']
            border = COLORS['byzantineBorder'] if is_byzantine else COLORS['honestBorder']
            ax.add_patch(plt.Rectangle((x - 0.5, y - 0.4), 1.0, 0.8,
                                       color=color, ec=border, lw=1))
            ax.text(x, y, f"V{v}", ha="center", va="center", fontsize=7, color="white")
 
# --- figure ---
fig, (ax, ax2) = plt.subplots(
    2, 1,
    figsize=(min(1 + len(rounds) * 3, 15), min(1 + n_vals * 1.5, 8.5)),
    gridspec_kw={'height_ratios': [4.5, 1]},
    sharex=True
)
fig.patch.set_facecolor("#f8f8f6")
ax.set_facecolor("#f8f8f6")
ax2.set_facecolor("#f8f8f6")
 
# top plot
draw_shading(ax)
draw_edges(ax)
draw_blocks(ax)

for r in rounds:
    ax.text((r - 1) * X_STEP, 0.9, f"Round {r}",
            ha="center", va="bottom", fontsize=10, color="#333", fontweight="500")
for v in validators:
    ax.text(-1.1, -v * Y_STEP, f"V{v}", ha="center", va="center", fontsize=9, color="#555")

ax.set_xlim(-1.5, (max(rounds) - 1) * X_STEP + 1.5)
ax.set_ylim(-(n_vals - 1) * Y_STEP - 0.6, 1.3)
ax.set_aspect("equal")
ax.axis("off")
ax.set_title("DAG Consensus Results", fontsize=13, pad=12)

# legend
ax2.legend(handles=[
    mpatches.Patch(color=COLORS['committed'], label="Committed block"),
    mpatches.Patch(color=COLORS['honest'],    label="Honest validator"),
    mpatches.Patch(color=COLORS['byzantine'], label="Byzantine validator"),
], loc="lower center", bbox_to_anchor=(0.5, -0.35), ncol=3, fontsize=8, framealpha=0.9)

# bottom plot
draw_byzantine_grid(ax2)

ax2.set_xlim(-1.5, (max(rounds) - 1) * X_STEP + 1.5)
ax2.set_ylim(-n_vals + 0.4, 0.6)
ax2.axis("off")
ax2.set_title("Validator Byzantine Status", fontsize=10)

plt.tight_layout()
plt.savefig("dag_visualization.png", dpi=150, bbox_inches="tight")
plt.show()