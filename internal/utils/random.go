package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var randomAdjectives = []string{
	"brave",
	"bright",
	"calm",
	"clever",
	"dapper",
	"eager",
	"gleaming",
	"gentle",
	"keen",
	"lively",
	"merry",
	"nimble",
	"proud",
	"spry",
	"steady",
	"swift",
	"valiant",
	"vibrant",
	"whimsical",
	"witty",
}

var randomNouns = []string{
	"otter",
	"panda",
	"quill",
	"raven",
	"pepper",
	"falcon",
	"harbor",
	"ivy",
	"jewel",
	"kestrel",
	"lantern",
	"meadow",
	"ocean",
	"prairie",
	"quartz",
	"rifle",
	"spruce",
	"thistle",
	"tide",
	"willow",
}

// RandomBranchName returns a Docker-style adjective-noun name for temporary branches.
func RandomBranchName() string {
	return fmt.Sprintf("%s-%s", randomWord(randomAdjectives), randomWord(randomNouns))
}

func randomWord(list []string) string {
	if len(list) == 0 {
		return ""
	}
	limit := big.NewInt(int64(len(list)))
	idx, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return list[0]
	}
	return list[int(idx.Int64())]
}
