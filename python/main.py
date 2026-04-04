import pygame
import sys
from validator import Validator


def main():
    # initialize pygame window
    pygame.init()
    screen = pygame.display.set_mode((1200, 800))
    pygame.display.set_caption("DAG-Based Consensus!")
    

    clock = pygame.time.Clock()
    font = pygame.font.SysFont("monospace", 18)
    paused = False

    # dag_sim = DAGSimulation()
    # pbft_sim = PBFTSimulation()

    while True:
        paused = handle_events(paused)

        screen.fill((10, 20, 30))

        # commands label 
        if paused:
            label = font.render("PAUSED  [SPACE to resume]", True, (255, 165, 0))
        else:
            label = font.render("RUNNING  [SPACE to pause]  [R to reset]", True, (0, 200, 100))
        screen.blit(label, (20, 760))

        # update entire window each frame
        pygame.display.flip()
        clock.tick(60)


def draw_dag(screen, sim):
    pass


def draw_pbft(screen, sim):
    pass


def handle_events(paused):
    """
    Process any pending pygame events (quit/pause/restart). 
    Returns the updated paused state of the simulation.
    """
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            pygame.quit()
            sys.exit()
        if event.type == pygame.KEYDOWN:
            if event.key == pygame.K_SPACE:
                paused = not paused
            if event.key == pygame.K_r:
                pass # TODO: reset both simulations
    return paused
    
class Simulation:
    def __init__(self, f=1):
        self.f = f
        self.n = 3*f + 1
        self.current_round = 0
        self.shared_chain = {}      # single shared chain
        self.validators = [Validator(i, f) for i in range(self.n)]

    # Progresses 1 round (when we do this in pygame we would call this at some interval)
    def tick(self):
        # Each validator proposes a vertex
        for v in self.validators:
            vertex = v.propose(self.current_round, self.shared_chain)
        
        # num_approvals = 0
        # for v in self.validators:
            if v.validate(vertex, self.shared_chain):
                self.shared_chain[vertex.vertex_id] = vertex

        # Each validator tries to commit older rounds
        for v in self.validators:
            v.try_commit(self.current_round, self.shared_chain)

        self.current_round += 1
        
if __name__ == "__main__":
    # NEED TO READD THE PYGAME MAIN STUFF
    sim = Simulation(f=1)

    for round_num in range(11):     # 10 simulated roundss
        sim.tick()
        print(f"Round {sim.current_round - 1}: "
            f"{len(sim.shared_chain)} total vertices, ")

    print("Final total order (V0):", sim.validators[0].total_order)
    print("All validators agree:", all(
        v.total_order == sim.validators[0].total_order for v in sim.validators
    ))
