package timing


import (
	"math"
)

var UseFractionalDelays bool
var UpdateFreq float64
var ShortestDelay struct {
	Lo, Hi int
}
var LongestDelay int
var SupportedLengths = []int{32, 24, 16, 12, 8, 6, 4, 3, 2, 1}


// Update the current minimum and maximum delay.
func UpdateDelayMinMax(delay int) {
	if (delay & 0xFF) < ShortestDelay.Lo &&
	   (delay & 0xFF) > 0 {
		ShortestDelay.Lo = delay & 0xFF
	}
	ShortestDelay.Hi = int(math.Min(float64(delay >> 8), float64(ShortestDelay.Hi)))
	LongestDelay = int(math.Max(float64(delay >> 8), float64(LongestDelay)))
}
