package bdd

import (
	"math/rand"
	"strconv"
)

var lastIdGen = 0
var nodeIdPrefix string
var nodeIdNchars = 6

func init() {
	ans := ""
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	nChars := len(chars)
	for i := 0; i < nodeIdNchars; i++ {
		ans += string(chars[rand.Intn(nChars)])
	}
	nodeIdPrefix = ans
}

func NextNodeId() string {
	ans := "node_" + nodeIdPrefix + "_" + strconv.Itoa(lastIdGen)
	lastIdGen++
	return ans
}
