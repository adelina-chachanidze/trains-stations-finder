package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "Error: incorrect number of arguments.")
		os.Exit(1)
	}

	mapPath, start, end, numTrainsStr := os.Args[1], os.Args[2], os.Args[3], os.Args[4]
	numTrains, err := parsePositiveInt(numTrainsStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: number of trains is not a valid positive integer.")
		os.Exit(1)
	}

	network, err := ParseNetwork(mapPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if start == end {
		fmt.Fprintln(os.Stderr, "Error: start and end station are the same.")
		os.Exit(1)
	}

	if !network.HasStation(start) {
		fmt.Fprintln(os.Stderr, "Error: start station does not exist.")
		os.Exit(1)
	}
	if !network.HasStation(end) {
		fmt.Fprintln(os.Stderr, "Error: end station does not exist.")
		os.Exit(1)
	}

	moves, err := FindTrainMovements(network, start, end, numTrains)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	for _, turn := range moves {
		fmt.Println(turn)
	}
}

func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, errors.New("not a valid positive integer")
	}
	return n, nil
}

func FindTrainMovements(network *Network, start, end string, numTrains int) ([]string, error) {
	// 1. Find all shortest paths from start to end using BFS
	paths := findAllShortestPaths(network, start, end)
	if len(paths) == 0 {
		return nil, errors.New("no path between the start and end stations")
	}

	// 2. Assign trains to paths in round-robin fashion
	trainPaths := make([][]string, numTrains)
	for i := 0; i < numTrains; i++ {
		trainPaths[i] = paths[i%len(paths)]
	}

	// 3. Simulate train movements turn by turn
	// Each train's position is an index in its path
	positions := make([]int, numTrains) // index in path for each train
	arrived := make([]bool, numTrains)
	turns := []string{}

	// To track station occupancy (except start/end)
	stationOccupied := make(map[string]int) // station name -> train index
	// To track track usage per turn
	for {
		moveThisTurn := []string{}
		trackUsed := make(map[string]bool)
		stationNextOccupied := make(map[string]int)
		progress := false
		for i := 0; i < numTrains; i++ {
			if arrived[i] {
				continue
			}
			currIdx := positions[i]
			path := trainPaths[i]
			if currIdx == len(path)-1 {
				arrived[i] = true
				continue
			}
			nextStation := path[currIdx+1]
			// Check if next station is occupied (except end)
			if nextStation != end {
				if _, occ := stationOccupied[nextStation]; occ {
					continue
				}
				if _, occ := stationNextOccupied[nextStation]; occ {
					continue
				}
			}
			// Check if track is used this turn (order stations to avoid directionality)
			trackKey := path[currIdx] + ":" + nextStation
			if trackUsed[trackKey] {
				continue
			}
			// Move train
			positions[i]++
			moveThisTurn = append(moveThisTurn, fmt.Sprintf("T%d-%s", i+1, nextStation))
			trackUsed[trackKey] = true
			if nextStation != end {
				stationNextOccupied[nextStation] = i
			}
			progress = true
		}
		if len(moveThisTurn) > 0 {
			turns = append(turns, strings.Join(moveThisTurn, " "))
		}
		// Update stationOccupied for next turn
		stationOccupied = stationNextOccupied
		if !progress {
			break
		}
	}
	return turns, nil
}

// Helper: BFS to find all shortest paths from start to end
func findAllShortestPaths(network *Network, start, end string) [][]string {
	type node struct {
		path    []string
		visited map[string]bool
	}
	var results [][]string
	minLen := -1
	queue := []node{{path: []string{start}, visited: map[string]bool{start: true}}}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		last := n.path[len(n.path)-1]
		if last == end {
			if minLen == -1 || len(n.path) == minLen {
				results = append(results, append([]string{}, n.path...))
				minLen = len(n.path)
			}
			continue
		}
		if minLen != -1 && len(n.path) >= minLen {
			continue
		}
		for neighbor := range network.Connections[last] {
			if n.visited[neighbor] {
				continue
			}
			newVisited := make(map[string]bool)
			for k, v := range n.visited {
				newVisited[k] = v
			}
			newVisited[neighbor] = true
			queue = append(queue, node{path: append(append([]string{}, n.path...), neighbor), visited: newVisited})
		}
	}
	return results
}
