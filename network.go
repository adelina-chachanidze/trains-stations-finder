package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Station struct {
	Name string
	X, Y int
}

type Network struct {
	Stations    map[string]*Station
	Coords      map[string]string              // "x,y" -> station name
	Connections map[string]map[string]struct{} // adjacency list
}

var stationNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

func ParseNetwork(path string) (*Network, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("could not open network map file")
	}
	defer file.Close()

	net := &Network{
		Stations:    make(map[string]*Station),
		Coords:      make(map[string]string),
		Connections: make(map[string]map[string]struct{}),
	}

	scanner := bufio.NewScanner(file)
	section := ""
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "stations:" {
			section = "stations"
			continue
		}
		if line == "connections:" {
			section = "connections"
			continue
		}
		if section == "stations" {
			parts := strings.Split(line, ",")
			if len(parts) != 1 && len(parts) != 3 {
				return nil, fmt.Errorf("invalid station line at %d", lineNum)
			}
			name := strings.TrimSpace(parts[0])
			if !stationNameRe.MatchString(name) {
				return nil, fmt.Errorf("invalid station name: %s", name)
			}
			if _, exists := net.Stations[name]; exists {
				return nil, fmt.Errorf("duplicate station name: %s", name)
			}
			station := &Station{Name: name}
			if len(parts) == 3 {
				xStr := strings.TrimSpace(parts[1])
				yStr := strings.TrimSpace(parts[2])
				x, err1 := strconv.Atoi(xStr)
				y, err2 := strconv.Atoi(yStr)
				if err1 != nil || err2 != nil || x <= 0 || y <= 0 {
					return nil, fmt.Errorf("invalid coordinates for station: %s", name)
				}
				coordKey := fmt.Sprintf("%d,%d", x, y)
				if other, exists := net.Coords[coordKey]; exists {
					return nil, fmt.Errorf("two stations at same coordinates: %s and %s", name, other)
				}
				station.X = x
				station.Y = y
				net.Coords[coordKey] = name
			}
			net.Stations[name] = station
			if _, ok := net.Connections[name]; !ok {
				net.Connections[name] = make(map[string]struct{})
			}
			continue
		}
		if section == "connections" {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid connection line at %d", lineNum)
			}
			a := strings.TrimSpace(parts[0])
			b := strings.TrimSpace(parts[1])
			if a == b {
				return nil, fmt.Errorf("connection from station to itself: %s", a)
			}
			if _, ok := net.Stations[a]; !ok {
				return nil, fmt.Errorf("connection references non-existent station: %s", a)
			}
			if _, ok := net.Stations[b]; !ok {
				return nil, fmt.Errorf("connection references non-existent station: %s", b)
			}
			// Check for duplicate or reverse duplicate
			if _, ok := net.Connections[a][b]; ok {
				return nil, fmt.Errorf("duplicate connection: %s-%s", a, b)
			}
			if _, ok := net.Connections[b][a]; ok {
				return nil, fmt.Errorf("duplicate connection: %s-%s", b, a)
			}
			net.Connections[a][b] = struct{}{}
			net.Connections[b][a] = struct{}{}
			continue
		}
		return nil, fmt.Errorf("unexpected line outside of section at %d", lineNum)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return net, nil
}

func (n *Network) HasStation(name string) bool {
	_, ok := n.Stations[name]
	return ok
}
