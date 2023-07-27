package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	dir := flag.String("dir", "",
		"path to a folder that contains simulation results")
	flag.Parse()

	if *dir == "" {
		panic("--dir can't be empty")
	}

	files, err := os.ReadDir(*dir)
	if err != nil {
		panic(err)
	}

	statsByDrone := map[string]*droneStats{}
	statsByBuild := map[string]*buildStats{}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(*dir, f.Name()))
		if err != nil {
			panic(err)
		}
		var results runResults
		if err := json.Unmarshal(data, &results); err != nil {
			panic(err)
		}
		keyParts := make([]string, 0, len(results.Drones))
		for _, drone := range results.Drones {
			keyParts = append(keyParts, drone)
			stats := statsByDrone[drone]
			if stats == nil {
				stats = &droneStats{name: drone}
				statsByDrone[drone] = stats
			}
			stats.picks++
			if results.Victory {
				stats.wins++
			}
		}
		sort.Strings(keyParts)
		buildKey := strings.Join(keyParts, ", ")
		stats := statsByBuild[buildKey]
		if stats == nil {
			stats = &buildStats{
				drones: results.Drones,
			}
			statsByBuild[buildKey] = stats
		}
		stats.picks++
		if results.Victory {
			stats.wins++
		}
	}

	var droneStatsList []*droneStats
	for _, stats := range statsByDrone {
		stats.winRate = float64(stats.wins) / float64(stats.picks)
		droneStatsList = append(droneStatsList, stats)
	}

	var buildStatsList []*buildStats
	for _, stats := range statsByBuild {
		stats.winRate = float64(stats.wins) / float64(stats.picks)
		buildStatsList = append(buildStatsList, stats)
	}

	sort.SliceStable(droneStatsList, func(i, j int) bool {
		return droneStatsList[i].winRate > droneStatsList[j].winRate
	})
	sort.SliceStable(buildStatsList, func(i, j int) bool {
		return buildStatsList[i].winRate > buildStatsList[j].winRate
	})

	for _, stats := range droneStatsList {
		fmt.Printf("%s => %d%%\n", stats.name, int(math.Round(100*stats.winRate)))
	}
	fmt.Println("----")
	for _, stats := range buildStatsList {
		if stats.picks < 3 {
			continue
		}
		fmt.Printf("%v => %d%% (%d picks)\n", stats.drones, int(math.Round(100*stats.winRate)), stats.picks)
	}
}

type buildStats struct {
	drones []string

	picks int
	wins  int

	winRate float64
}

type droneStats struct {
	name string

	picks int
	wins  int

	winRate float64
}

type runResults struct {
	Seed    int
	Env     int
	Victory bool
	Score   int
	Time    int
	Mode    string

	Drones []string
	Turret string
	Core   string
}
