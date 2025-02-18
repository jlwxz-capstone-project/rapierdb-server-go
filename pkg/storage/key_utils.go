package storage

import (
	"strings"
)

// CalcDocKey 计算文档的键值
// 键值格式为 "d<dbName>:<collectionName>:<docID>"，每个字段都填充到固定长度
// 如果 dbName、collectionName 或 docID 超过最大长度限制，则返回错误
func CalcDocKey(dbName, collectionName, docID string) ([]byte, error) {
	// 先检查长度限制
	if len(dbName) > DB_SIZE_IN_BYTES {
		return nil, ErrDbNameTooLarge
	}
	if len(collectionName) > COLLECTION_SIZE_IN_BYTES {
		return nil, ErrCollectionNameTooLarge
	}
	if len(docID) > DOC_ID_SIZE_IN_BYTES {
		return nil, ErrDocIDTooLarge
	}

	result := make([]byte, KEY_SIZE_IN_BYTES)
	n := copy(result, DOC_KEY_PREFIX)

	// 填充数据库名称
	dbNameLen := copy(result[n:], dbName)
	for i := n + dbNameLen; i < n+DB_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += DB_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// 填充集合名称
	collectionNameLen := copy(result[n:], collectionName)
	for i := n + collectionNameLen; i < n+COLLECTION_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += COLLECTION_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// 填充文档ID
	docIDLen := copy(result[n:], docID)
	for i := n + docIDLen; i < n+DOC_ID_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}

	return result, nil
}

// CalcCollectionLowerBound 计算集合键值范围的下界
// 下界格式为 "d<dbName>:<collectionName>:"，每个字段都填充到固定长度，后面用0填充
// 如果 dbName 或 collectionName 超过最大长度限制，则返回错误
func CalcCollectionLowerBound(dbName, collectionName string) ([]byte, error) {
	result := make([]byte, KEY_SIZE_IN_BYTES)
	n := copy(result, DOC_KEY_PREFIX)

	// 填充数据库名称
	dbNameLen := copy(result[n:], dbName)
	if dbNameLen > DB_SIZE_IN_BYTES {
		return nil, ErrDbNameTooLarge
	}
	for i := n + dbNameLen; i < n+DB_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += DB_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// 填充集合名称
	collectionNameLen := copy(result[n:], collectionName)
	if collectionNameLen > COLLECTION_SIZE_IN_BYTES {
		return nil, ErrCollectionNameTooLarge
	}
	for i := n + collectionNameLen; i < n+COLLECTION_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += COLLECTION_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// 填充文档ID部分为0
	for i := n; i < KEY_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}

	return result, nil
}

// CalcCollectionUpperBound 计算集合键值范围的上界
// 上界格式为 "d<dbName>:<collectionName>:"，每个字段都填充到固定长度，后面用0xFF填充
// 如果 dbName 或 collectionName 超过最大长度限制，则返回错误
func CalcCollectionUpperBound(dbName, collectionName string) ([]byte, error) {
	result := make([]byte, KEY_SIZE_IN_BYTES)
	n := copy(result, DOC_KEY_PREFIX)

	// 填充数据库名称
	dbNameLen := copy(result[n:], dbName)
	if dbNameLen > DB_SIZE_IN_BYTES {
		return nil, ErrDbNameTooLarge
	}
	for i := n + dbNameLen; i < n+DB_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += DB_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// 填充集合名称
	collectionNameLen := copy(result[n:], collectionName)
	if collectionNameLen > COLLECTION_SIZE_IN_BYTES {
		return nil, ErrCollectionNameTooLarge
	}
	for i := n + collectionNameLen; i < n+COLLECTION_SIZE_IN_BYTES; i++ {
		result[i] = 0
	}
	n += COLLECTION_SIZE_IN_BYTES
	result[n] = ':'
	n++

	// 填充文档ID部分为0xFF
	for i := n; i < KEY_SIZE_IN_BYTES; i++ {
		result[i] = 0xFF
	}

	return result, nil
}

func GetDatabaseNameFromKey(key []byte) string {
	n := len(DOC_KEY_PREFIX)
	// 去除尾部的空字节
	dbName := string(key[n : n+DB_SIZE_IN_BYTES])
	return strings.TrimRight(dbName, "\x00")
}

func GetCollectionNameFromKey(key []byte) string {
	n := len(DOC_KEY_PREFIX) + DB_SIZE_IN_BYTES + 1 // +1 跳过冒号
	// 去除尾部的空字节
	collectionName := string(key[n : n+COLLECTION_SIZE_IN_BYTES])
	return strings.TrimRight(collectionName, "\x00")
}

func GetDocIdFromKey(key []byte) string {
	n := len(DOC_KEY_PREFIX) + DB_SIZE_IN_BYTES + 1 + COLLECTION_SIZE_IN_BYTES + 1 // +1 跳过两个冒号
	// 去除尾部的空字节
	docId := string(key[n : n+DOC_ID_SIZE_IN_BYTES])
	return strings.TrimRight(docId, "\x00")
}
