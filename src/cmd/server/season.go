package main

const currentSeason = 0

func seasonByBuild(version int) int {
	if version <= 13 {
		return 0
	}

	return -1
}
