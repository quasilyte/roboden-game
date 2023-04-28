package main

const currentSeason = 0

func seasonByBuild(version int) int {
	if version <= 11 {
		return 0
	}

	return -1
}
