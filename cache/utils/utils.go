package utils

func SumMap64(mp map[string]int64) (result int64) {
	for _, value := range mp {
		result += value
	}
	return
}
