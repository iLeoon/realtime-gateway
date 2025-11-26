package websocket

import "errors"

var (
	ErrUpgradeConnection = errors.New("error on upgrading raw tcp connection into websocekt")
)

