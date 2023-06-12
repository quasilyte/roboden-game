package main

const currentSeason = 0

func seasonByBuild(version int) int {
	if version <= 12 {
		return 0
	}

	return -1
}
