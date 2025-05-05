package key_utils

import (
	"strings"

	pe "github.com/pkg/errors"
)

const (
	STORAGE_META_KEY         = "m" // Key for storing metadata
	DOC_KEY_PREFIX           = "d" // Prefix for document keys
	COLLECTION_SIZE_IN_BYTES = 16  // Maximum bytes for collection name
	DOC_ID_SIZE_IN_BYTES     = 16  // Maximum bytes for document ID
)

// Total bytes for a document key = prefix(1) + collection name bytes(16) + sep(1) + doc id bytes(16)
var KEY_SIZE_IN_BYTES = 1 + COLLECTION_SIZE_IN_BYTES + 1 + DOC_ID_SIZE_IN_BYTES

// CalcDocKey calculates the key for a document.
// Key format is "d<collectionName>:<docID>", each field is padded to a fixed length.
// Returns an error if collectionName or docID exceeds the maximum length limit.
func CalcDocKey(collectionName, docID string) ([]byte, error) {
	// Check length limits first
	if len(collectionName) > COLLECTION_SIZE_IN_BYTES {
		return nil, pe.Errorf("collection name too large: %s", collectionName)
	}
	if len(docID) > DOC_ID_SIZE_IN_BYTES {
		return nil, pe.Errorf("doc id too large: %s", docID)
	}

	result := make([]byte, KEY_SIZE_IN_BYTES)
	n := copy(result, DOC_KEY_PREFIX)

	// Pad collection name
	collectionNameLen := copy(result[n:], collectionName)
	for i := n + collectionNameLen; i < n+COLLECTION_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += COLLECTION_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// Pad document ID
	docIDLen := copy(result[n:], docID)
	for i := n + docIDLen; i < n+DOC_ID_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}

	return result, nil
}

// CalcCollectionLowerBound calculates the lower bound of the key range for a collection.
// Lower bound format is "d<collectionName>:", each field is padded to a fixed length, followed by 0s.
// Returns an error if collectionName exceeds the maximum length limit.
func CalcCollectionLowerBound(collectionName string) ([]byte, error) {
	result := make([]byte, KEY_SIZE_IN_BYTES)
	n := copy(result, DOC_KEY_PREFIX)

	// Pad collection name
	collectionNameLen := copy(result[n:], collectionName)
	if collectionNameLen > COLLECTION_SIZE_IN_BYTES {
		return nil, pe.Errorf("collection name too large: %s", collectionName)
	}
	for i := n + collectionNameLen; i < n+COLLECTION_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += COLLECTION_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// Pad the document ID part with 0s
	for i := n; i < KEY_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}

	return result, nil
}

// CalcCollectionUpperBound calculates the upper bound of the key range for a collection.
// Upper bound format is "d<collectionName>:", each field is padded to a fixed length, followed by 0xFF.
// Returns an error if collectionName exceeds the maximum length limit.
func CalcCollectionUpperBound(collectionName string) ([]byte, error) {
	result := make([]byte, KEY_SIZE_IN_BYTES)
	n := copy(result, DOC_KEY_PREFIX)

	// Pad collection name
	collectionNameLen := copy(result[n:], collectionName)
	if collectionNameLen > COLLECTION_SIZE_IN_BYTES {
		return nil, pe.Errorf("collection name too large: %s", collectionName)
	}
	for i := n + collectionNameLen; i < n+COLLECTION_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += COLLECTION_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// Pad the document ID part with 0xFF
	for i := n; i < KEY_SIZE_IN_BYTES; i++ {
		result[i] = 0xFF
	}

	return result, nil
}

// GetCollectionNameFromKey extracts the collection name from a document key.
func GetCollectionNameFromKey(key []byte) string {
	n := len(DOC_KEY_PREFIX)
	// Remove trailing null bytes
	collectionName := string(key[n : n+COLLECTION_SIZE_IN_BYTES])
	return strings.TrimRight(collectionName, "\x00")
}

// GetDocIdFromKey extracts the document ID from a document key.
func GetDocIdFromKey(key []byte) string {
	n := len(DOC_KEY_PREFIX) + COLLECTION_SIZE_IN_BYTES + 1 // +1 to skip the colon
	// Remove trailing null bytes
	docId := string(key[n : n+DOC_ID_SIZE_IN_BYTES])
	return strings.TrimRight(docId, "\x00")
}
