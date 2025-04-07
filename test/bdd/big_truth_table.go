package main

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
)

// BigTruthTable 示例大型真值表
var BigTruthTable = map[string]int{
	"10000010010000001": 1,
	"10011010010000001": 1,
	"10010010010000001": 1,
	"10000000010000101": 2,
	"10000000010000001": 10,
	"10010000010000101": 2,
	"10010000010000001": 10,
	"01000000001000011": 8,
	"01010000101000011": 8,
	"01010000001000011": 8,
	"10010000110000001": 12,
	"10000000010001001": 1,
	"10011000110001001": 5,
	"10010000110001001": 6,
	"10010000010001001": 1,
	"10010110010000001": 12,
	"01011000101000011": 5,
	"10010100010001001": 12,
	"01000000000010101": 2,
	"01000000000000001": 12,
	"01010000100000001": 12,
	"01010000000010101": 2,
	"01010000000000001": 12,
	"01010100001000011": 8,
	"10000110010000001": 12,
	"10011110010000001": 12,
	"01000000011000011": 12,
	"01010000111000111": 12,
	"01010000111000011": 12,
	"01010000011000011": 12,
	"01011000111000111": 12,
	"01010100110100011": 12,
	"10000100010001001": 12,
	"10000100010000001": 10,
	"10011100110001001": 12,
	"10010100010000001": 10,
	"10010100110001001": 12,
	"01000100010100011": 12,
	"01011100110100111": 12,
	"01010100010100011": 12,
	"01000000010100001": 12,
	"01000000010000001": 12,
	"01010000110000001": 12,
	"01010000010100001": 12,
	"01010000010000001": 12,
	"01000100000101001": 12,
	"01000100010101001": 12,
	"01000100010000001": 12,
	"01011100100101001": 12,
	"01011100110101001": 12,
	"01011100110000001": 12,
	"01010100000101001": 12,
	"01010100010101001": 12,
	"01010100010000001": 12,
	"01010100110100111": 12,
	"01000100001000011": 8,
	"01010100101000011": 8,
	"01000000000101001": 1,
	"01011000100101001": 5,
	"01010000100101001": 6,
	"01010000000101001": 1,
	"10010100110000001": 12,
	"01000100011000011": 12,
	"01010100111000011": 12,
	"01010100011000011": 12,
	"01011100101000011": 5,
	"01000000001000010": 7,
	"01011000101000010": 12,
	"01010000101000010": 12,
	"01010000001000010": 7,
	"10000100010000101": 2,
	"10010100010000101": 2,
	"01010000110010011": 12,
	"01000100011001011": 12,
	"01011100110011011": 12,
	"01010100110011011": 12,
	"01010100011001011": 12,
	"01000000011001011": 11,
	"01011000110011011": 5,
	"01010000110011011": 6,
	"01010000011001011": 11,
	"01011100111001011": 12,
	"01000100000010101": 2,
	"01000100000000001": 12,
	"01011100100000001": 12,
	"01010100100101001": 12,
	"01010100100000001": 12,
	"01010100000010101": 2,
	"01010100000000001": 12,
	"01000100001000010": 7,
	"01011100101000010": 12,
	"01010100101000010": 12,
	"01010100001000010": 7,
	"01000000011000111": 11,
	"01000100011000111": 11,
	"01010100111000111": 12,
	"01000100010100111": 5,
	"01010100010100111": 5,
	"01010100011000111": 11,
	"01000000010100101": 2,
	"01010000010100101": 2,
	"01011000110000001": 12,
	"01011100111000111": 12,
	"01010000110100001": 12,
	"01010000011000111": 11,
	"01000100010100001": 12,
	"01010100010100001": 12,
	"01010100110101001": 12,
	"01010100110000001": 12,
	"01000110000000001": 12,
	"01000110010000001": 12,
	"01011110000000001": 12,
	"01011110010000001": 12,
	"01010110000000001": 12,
	"01010110010000001": 12,
	"01011100111000011": 5,
	"01010100110100001": 12,
	"01000000010101001": 1,
	"01011000110101001": 5,
	"01010000110101001": 6,
	"01010000010101001": 1,
	"01000000011000010": 7,
	"01011000111000110": 12,
	"01010000111000110": 12,
	"01010000011000010": 7,
	"01000100010100010": 3,
	"01000100011000010": 7,
	"01011100110100110": 12,
	"01010100110100110": 12,
	"01010100111000010": 12,
	"01010100110100010": 12,
	"01010100010100010": 3,
	"01010100011000010": 7,
	"01011000111000011": 5,
	"00100000001000010": 7,
	"00110000101000010": 12,
	"00100100001000010": 7,
	"00111100101000010": 12,
	"00110100101000010": 12,
	"00100100000100010": 3,
	"00100100000000010": 3,
	"00111100100100010": 12,
	"00111100100000010": 12,
	"00110100100100010": 12,
	"00110100100000010": 12,
	"00110100001000010": 7,
	"00110100000100010": 3,
	"00110100000000010": 3,
	"00110000001000010": 7,
	"00111000101000010": 12,
	"01000001011000001": 12,
	"01000001011001001": 11,
	"01011001111000001": 5,
	"01011001111001001": 5,
	"01010001011000001": 12,
	"01010001011001001": 11,
	"01000001011000000": 7,
	"01000001011001000": 7,
	"01011001111000000": 12,
	"01011001111001000": 12,
	"01010001011000000": 7,
	"01010001011001000": 7,
	"01010001111000001": 12,
	"01011001111000101": 12,
	"01010001111000101": 12,
	"01000001010000001": 12,
	"01011001110000001": 12,
	"01010001110000001": 12,
	"01010001010000001": 12,
	"01000101011000001": 12,
	"01011101111000001": 5,
	"01010101011000001": 12,
	"01010101111000001": 12,
	"01000001010000101": 2,
	"01000001010001001": 1,
	"01011001110001001": 5,
	"01010001010000101": 2,
	"01010001010001001": 1,
	"01000101011001001": 12,
	"01011101110001001": 12,
	"01010101111001001": 12,
	"01010101011001001": 12,
	"01010101110001001": 12,
	"01010001111000000": 12,
	"01000001011000101": 11,
	"01010001011000101": 11,
	"01000101010000101": 12,
	"01011101110000101": 12,
	"01010101110000101": 12,
	"01010101010000101": 12,
	"01010001111001000": 12,
	"01010101111000101": 12,
	"01000101010001001": 12,
	"01000101010000001": 12,
	"01011101110000001": 12,
	"01010101110000001": 12,
	"01010101010001001": 12,
	"01010101010000001": 12,
	"01000101011000101": 11,
	"01010101011000101": 11,
	"01011001111000100": 12,
	"01010001111000100": 12,
	"01000101011000000": 7,
	"01000101010000000": 12,
	"01011101110000100": 12,
	"01010101110000100": 12,
	"01010101111000000": 12,
	"01010101011000000": 7,
	"01010101010000000": 12,
	"01000111010000001": 12,
	"01011111010000001": 12,
	"01010111010000001": 12,
	"01010001110001001": 6,
	"00100001001000000": 7,
	"00110001101000000": 12,
	"00100101001000000": 7,
	"00100101000000000": 12,
	"00111101100000000": 12,
	"00110101100000000": 12,
	"00110101101000000": 12,
	"00110101000000000": 12,
	"00111001101000000": 12,
	"00110001001000000": 7,
	"00110101001000000": 7,
	"00111101101000000": 12,
	"01010000111001011": 11,
	"01000000011001010": 7,
	"01010000111001010": 12,
	"01010000111000010": 12,
	"01010000011001010": 7,
	"01011000111001011": 5,
	"01010001111001001": 11,
	"01000010000000001": 1,
	"01000010010000001": 1,
	"01011010000000001": 1,
	"01011010010000001": 1,
	"01010010000000001": 1,
	"01010010010000001": 1,
	"01000011010000001": 1,
	"01011011010000001": 1,
	"01010011010000001": 1,
	"01011000111001010": 12,
	"01011000111000010": 12,
	"01011000110010011": 12,
	"01010000110000011": 12,
	"01011000100000001": 12,
	"01010101010001000": 12,
	"01010101110000000": 12,
	"10000010010000000": 0,
	"10011010010000000": 0,
	"10010010010000000": 0,
	"10000110010000000": 0,
	"10011110010000000": 0,
	"10010110010000000": 0,
	"01000110000000011": 0,
	"01011110000000011": 0,
	"01010110000000011": 0,
	"01000010000000000": 0,
	"01011010000000000": 0,
	"01010010000000000": 0,
	"01000110000000000": 0,
	"01011110000000000": 0,
	"01010110000000000": 0,
	"01000110000000010": 0,
	"01011110000000010": 0,
	"01010110000000010": 0,
	"10011000110000101": 0,
	"10011000110000001": 0,
	"01011000100010111": 0,
	"01011000100000011": 0,
	"10010000110000101": 0,
	"01011000110000011": 0,
	"01010000100010111": 0,
	"01000110010000011": 0,
	"01011110010000011": 0,
	"01010110010000011": 0,
	"01000010010000000": 0,
	"01011010010000000": 0,
	"01010010010000000": 0,
	"01000110010000000": 0,
	"01011110010000000": 0,
	"01010110010000000": 0,
	"01000000000010100": 0,
	"01000000000000000": 0,
	"01011000100010100": 0,
	"01011000100000000": 0,
	"01010000000010100": 0,
	"01010000000000000": 0,
	"10000000010000100": 0,
	"10000000010001000": 0,
	"10000000010000000": 0,
	"10011000110000100": 0,
	"10011000110001000": 0,
	"10011000110000000": 0,
	"10010000010000100": 0,
	"10010000010001000": 0,
	"10010000010000000": 0,
	"10010000110000000": 0,
	"10010000110001000": 0,
	"01010000100000011": 0,
	"01010100000101011": 0,
	"01000000010100100": 0,
	"01000000010000000": 0,
	"01011000110100100": 0,
	"01011000110000000": 0,
	"01010000010100100": 0,
	"01010000010000000": 0,
	"01011000100010101": 0,
	"01011000110100101": 0,
	"01010000110100101": 0,
	"01011000110010111": 0,
	"01010000110010111": 0,
	"01000000010101000": 0,
	"01011000110101000": 0,
	"01010000010101000": 0,
	"01010000100010101": 0,
	"01010100010101011": 0,
	"01000000010100000": 0,
	"01010000100010100": 0,
	"01010000110100000": 0,
	"01010000110000000": 0,
	"01010000010100000": 0,
	"00111000100010010": 0,
	"00111000100000010": 0,
	"00110000100010010": 0,
	"00110000100000010": 0,
	"00100110000000010": 0,
	"00111110000000010": 0,
	"00110110000000010": 0,
	"00100010000000000": 0,
	"00111010000000000": 0,
	"00110010000000000": 0,
	"00100110000000000": 0,
	"00111110000000000": 0,
	"00110110000000000": 0,
	"00100000000010000": 0,
	"00100000000000000": 0,
	"00111000100010000": 0,
	"00111000100000000": 0,
	"00110000100010000": 0,
	"00110000100000000": 0,
	"00110000000010000": 0,
	"00110000000000000": 0,
	"00100000000100000": 0,
	"00111000100100000": 0,
	"00110000000100000": 0,
	"01000011010000000": 0,
	"01011011010000000": 0,
	"01010011010000000": 0,
	"01000111010000000": 0,
	"01011111010000000": 0,
	"01010111010000000": 0,
	"01011001110000101": 0,
	"01000001010001000": 0,
	"01000001010000000": 0,
	"01011001110001000": 0,
	"01011001110000000": 0,
	"01010001010001000": 0,
	"01010001010000000": 0,
	"01010001110000101": 0,
	"00111001100000000": 0,
	"00110001100000000": 0,
	"00100111000000000": 0,
	"00111111000000000": 0,
	"00110111000000000": 0,
	"00100011000000000": 0,
	"00111011000000000": 0,
	"00110011000000000": 0,
	"01000000000101000": 0,
	"01000000010001000": 0,
	"01011000100101000": 0,
	"01011000110011000": 0,
	"01010000100101000": 0,
	"01010000110001000": 0,
	"01010000000101000": 0,
	"01010000010001000": 0,
	"01011000110011010": 0,
	"01011000110000010": 0,
	"01010000110011010": 0,
	"01010000110000010": 0,
	"01000110010000010": 0,
	"01011110010000010": 0,
	"01010110010000010": 0,
	"00110000100100000": 0,
	"01000001010000100": 0,
	"01011001110000100": 0,
	"01010001010000100": 0,
	"00100001000000000": 0,
	"00110001000000000": 0,
	"01010001110000000": 0,
	"01010001110001000": 0,
	"01010000100000000": 0,
	"01000000010011000": 0,
	"01010000010011000": 0,
	"01000100000101011": 0,
	"01000100010101011": 0,
	"01000000010010100": 0,
	"01011000110010100": 0,
	"01010000010010100": 0,
	"10010000110000100": 0,
	"01010100010001011": 0,
	"01010001110000100": 0,
	"01011000100010110": 0,
	"01011000100000010": 0,
	"01010000100010110": 0,
	"01010000100000010": 0,
	"01010100000000011": 0,
	"10010100010000000": 0,
	"01010100110000011": 12,
	"00110100000000000": 0,
	"01010100010000011": 12,
	"10010100010001000": 0,
	"01010100010100101": 2,
	"01010100011001010": 7,
	"00110100000100000": 0,
	"01010101011001000": 7,
	"01010100010000010": 3,
	"10000100010000000": 0,
	"01000100010000011": 12,
	"00100100000000000": 0,
	"01000100000000011": 0,
	"01000100010000010": 3,
	"01011100100101011": 0,
	"01011100100000011": 0,
	"00110100100000000": 0,
	"01010100100000011": 0,
	"01010100000101000": 0,
	"00100100000010000": 0,
	"10000100010000100": 0,
	"01000101010000100": 0,
	"10000100010001000": 0,
	"01000101010001000": 12,
	"01000100000000000": 0,
	"01010100000101010": 3,
	"01000100000101010": 3,
	"01000000010000100": 0,
	"01010000110000100": 0,
	"01010000010000100": 0,
	"01000000010010000": 0,
	"01010000110010000": 0,
	"01010000010010000": 0,
	"01000100000000010": 3,
	"10011100110000001": 0,
	"01010100100101011": 0,
	"01010100000000000": 0,
	"10010100010000100": 0,
	"10011100110000101": 0,
	"10010100110000101": 0,
	"01000100000101000": 0,
	"01011100100101000": 0,
	"01011100100000000": 0,
	"01010100100000000": 0,
	"01010100000010100": 0,
	"00110100000010000": 0,
	"01010101010000100": 0,
	"01010000110101000": 0,
	"01010100000000010": 3,
	"01010000110010100": 0,
	"01000100010000000": 0,
	"01010100010000000": 0,
	"01000100010100101": 2,
	"01000100010101010": 3,
	"00100100000100000": 0,
	"01010100010101010": 3,
	"01011100100010111": 0,
	"01010100100010111": 0,
	"10010100110000000": 0,
	"00111100100010010": 0,
	"00111100100100000": 0,
	"01011100100101010": 12,
	"10011100110001000": 0,
	"01011101110001000": 12,
	"01011000110100001": 12,
	"01011000110001011": 5,
	"01011000110001010": 0,
	"01010100010101000": 0,
	"01010100010100100": 0,
	"01010100010011000": 0,
	"01000100011001010": 7,
	"01000100010001011": 0,
	"01011100100010101": 0,
	"01000101011001000": 7,
	"01000100010101000": 0,
	"01000100000010100": 0,
	"01000100010100100": 0,
	"01000100010011000": 0,
	"01010000110011000": 0,
	"01011100110000011": 0,
	"01010000110100100": 0,
	"00110100100010010": 0,
	"10010100110001000": 0,
	"01010100100101000": 0,
	"00110100100100000": 0,
	"01011100110000000": 0,
	"00111100100000000": 0,
	"01011100100010100": 0,
	"00111100100010000": 0,
	"10011100110000000": 0,
	"01010100010010100": 0,
	"10011100110000100": 0,
	"01011101110000000": 12,
	"01011101111001001": 12,
	"01011101111001000": 12,
	"01011100110010111": 0,
	"01010100110010111": 0,
	"01010100100101010": 12,
	"01011100110101011": 0,
	"01010100111001011": 12,
	"01011100110100101": 0,
	"01011100111001010": 12,
	"01011100110101010": 12,
	"01011000110010000": 0,
	"01010100010001000": 0,
	"01011100110001011": 12,
	"01000100010010100": 0,
	"01000100010001000": 0,
	"01010100010010000": 0,
	"01000100010010000": 0,
	"01011101111000101": 12,
	"01010100010000100": 0,
	"01000100010000100": 0,
	"01010100110100101": 0,
	"01010100111001010": 12,
	"01010101110001000": 12,
	"01010101111001000": 12,
	"01011000110000100": 0,
	"01010100110101011": 0,
	"01010100010100000": 0,
	"01000100010100000": 0,
	"01010100110010011": 12,
	"01011000110100000": 0,
	"01010000110001011": 6,
	"01010000110001010": 0,
	"01011100110101000": 0,
	"01011100110010100": 0,
	"01011100110001000": 0,
	"01011100110100100": 0,
	"01010100110001011": 0,
	"01011100100000010": 12,
	"01011100100010110": 0,
	"01011101111000000": 12,
	"01010100010000111": 5,
	"01000100010000111": 5,
	"01010100110000111": 12,
	"01011100110000111": 12,
	"01011100110011010": 0,
	"01010100110011010": 0,
	"01011100110011000": 0,
	"01010100110000010": 12,
	"01011100110100011": 5,
	"01010000110000111": 0,
	"01011100110000010": 12,
	"10010100110000100": 0,
	"01010100110010100": 0,
	"01010100110010000": 0,
	"01010100110000100": 0,
	"01010100110100000": 0,
	"01010100110100100": 0,
	"01010100110000000": 0,
	"01010100010001010": 3,
	"01011100110001010": 3,
	"01000100010001010": 3,
	"01010100100010110": 0,
	"01010100110101010": 12,
	"01010100100010100": 0,
	"01011100110100000": 0,
	"01011100111000010": 12,
	"01011000110000111": 0,
	"00110100100010000": 0,
	"01011000110001000": 0,
	"01010100100000010": 12,
	"01010100110001000": 0,
	"01010100110011000": 0,
	"01010100100010101": 0,
	"01010100110101000": 0,
	"01011100110100001": 0,
}

// GetBigTruthTable 获取大的真值表
func GetBigTruthTable() bdd.TruthTable {
	var firstKey string
	for k := range BigTruthTable {
		firstKey = k
		break
	}

	if firstKey != "" {
		keyLength := len(firstKey)
		bdd.FillTruthTable(BigTruthTable, keyLength, bdd.Unknown)
	}

	return BigTruthTable
}
