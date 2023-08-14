package main

const currentSeason = 1

func seasonByBuild(version int) int {
	switch {
	case version <= 13:
		return 0
	case version <= 15:
		return 1
	}

	return -1
}
