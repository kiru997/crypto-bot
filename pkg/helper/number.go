package helper

func PercentageChange(old, new float64) (delta float64) {
	diff := (new - old)
	delta = (diff / (old)) * 100
	return
}
