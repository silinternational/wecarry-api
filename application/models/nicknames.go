package models

import (
	"math/rand"
	"time"
)

func allPrefixes() [30]string {
	return [30]string{
		"Faithful",
		"Honest",
		"Determined",
		"Joyful",
		"Loving",
		"Peaceful",
		"Patient",
		"Kind",
		"Good",
		"Selfless",
		"Disciplined",
		"Authentic",
		"Powerful",
		"Enduring",
		"Trustworthy",
		"Upright",
		"Courageous",
		"Reliable",
		"Responsible",
		"Generous",
		"Compassionate",
		"Committed",
		"Humble",
		"Loyal",
		"Wise",
		"Truthful",
		"Diligent",
		"Flexible",
		"Calm",
		"Confident",
	}
}

func getShuffledPrefixes() [30]string {
	ps := allPrefixes()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(ps), func(i, j int) { ps[i], ps[j] = ps[j], ps[i] })
	return ps
}
