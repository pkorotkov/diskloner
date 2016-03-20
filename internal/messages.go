package internal

type Message interface {
	UUID() string
}

type CloningMessage struct {
	UUId        string
	CopiedBytes int64
	TotalBytes  int64
}

func (m *CloningMessage) UUID() string {
	return m.UUId
}

type InquiringMessage struct {
	UUId        string
	CopiedBytes int64
	TotalBytes  int64
}

func (m *InquiringMessage) UUID() string {
	return m.UUId
}

type CompletedMessage struct {
	UUId string
}

func (m *CompletedMessage) UUID() string {
	return m.UUId
}

type AbortedMessage struct {
	UUId string
}

func (m *AbortedMessage) UUID() string {
	return m.UUId
}
