import pygame
import sys


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


if __name__ == "__main__":
    main()