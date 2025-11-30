#!/usr/bin/env python3

"""
Principles:
- Everyone should get matched to someone else
- Should be able to weight or anti-weight
- Should be a little bit random (think different pairings for subsequent years)
"""

import collections
import random
from typing import Optional, Tuple

PEOPLE = ["Charlie", "Karen", "Anne", "Sam", "Bob"]


def random_pairings(people: list[str]) -> list[Tuple[str, str]]:
    copied = list(people)
    random.shuffle(copied)
    partners = collections.deque(copied)
    partners.rotate()

    return list(zip(copied, partners))


# Allow exclusions
# https://binary-machinery.github.io/2021/02/03/secret-santa-graph.html

PEOPLE_WITH_EXCLUSIONS = {
    "Charlie": [],
    "Karen": [],
    "Anne": [],
    "Bob": ["Sam"],
    "Sam": ["Bob"],
}


class Graph:
    def __init__(self, nodes: dict[str, set[str]]):
        self.nodes = nodes

    def pairings_new(self):
        def search(start: str, current_person: str, visited: set[str] = None) -> Optional[list[Tuple[str, str]]]:
            if visited is None:
                visited = set()

            next_candidates = [p for p in self.nodes[current_person] if p not in visited and p != start]
            random.shuffle(next_candidates)

            if len(visited) == len(self.nodes) - 1:
                if start in self.nodes[current_person]:
                    return [(current_person, start)]

                # If we've visited everyone except the starting person but the starting
                # person is not connected to the current person, this cycle is not
                # valid.
                return None

            for next_person in next_candidates:
                next_visited = visited.copy()
                next_visited.add(next_person)

                if (next_solution := search(start, next_person, next_visited)) is not None:
                    solution = [(current_person, next_person)]
                    solution.extend(next_solution)
                    return solution

            return None

        search_starts = list(self.nodes.keys())
        for top_level_start in search_starts:
            if (solution := search(top_level_start, top_level_start)) is not None:
                return solution




def main():
    random_outcome = random_pairings(PEOPLE_ALL)

    print("Random Pairings:")
    for gifter, giftee in random_outcome:
        print(f"  {gifter} -> {giftee}")

    nodes = {}
    for person, constraints in PEOPLE_WITH_EXCLUSIONS.items():
        edges = set(p for p in PEOPLE_WITH_EXCLUSIONS if p != person and p not in constraints)
        nodes[person] = edges

    graph = Graph(nodes)
    smart_outcome = graph.pairings_new()

    print("Smarter Pairings:")
    if smart_outcome is None:
        print("*** Smart Outcome was not solvable! ***")
    else:
        for gifter, giftee in smart_outcome:
            print(f"  {gifter} -> {giftee}")


if __name__ == "__main__":
    main()
