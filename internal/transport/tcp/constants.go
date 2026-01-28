package tcp

import "time"

const (
	writeDuration = 5 * time.Second
	readDuration  = 60 * time.Second
	pingTime      = 20 * time.Second
	pongWait      = 60 * time.Second
)
