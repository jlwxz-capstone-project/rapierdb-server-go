package bdd

import (
	"math/rand"
	"strconv"
)

var lastIdGen = 0
var nodePrefix string

const prefixLength = 6

func init() {
	res := ""
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	charsLen := len(chars)
	for i := 0; i < prefixLength; i++ {
		res += string(chars[rand.Intn(charsLen)])
	}
	nodePrefix = res
}

func NextNodeId() string {
	ret := "node_" + nodePrefix + "_" + strconv.Itoa(lastIdGen)
	lastIdGen++
	return ret
}
