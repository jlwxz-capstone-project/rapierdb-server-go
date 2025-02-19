package message

import "bytes"

// SyncMessage 由服务端发送给客户端
// 携带有客户端需要同步的数据
type SyncMessageV1 struct {
}

func (m *SyncMessageV1) Encode() ([]byte, error) {
	return nil, nil
}

func DecodeSyncMessageV1(b *bytes.Buffer) (*SyncMessageV1, error) {
	return nil, nil
}
