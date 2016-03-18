package internal

// TODO: Remove.
// var ProgressState struct {
// 	Aborted, Cloning, Inquiring, Completed string
// } = struct {
// 	Aborted, Cloning, Inquiring, Completed string
// }{
// 	"Aborted",
// 	"Cloning",
// 	"Inquiring",
// 	"Completed",
// }

type CloningMessage struct {
	// State       string
	Name        string
	UUID        string
	CopiedBytes int64
	TotalBytes  int64
}

type InquiringMessage struct {
	UUID string
}

type CompletedMessage struct {
	UUID string
}

type AbortedMessage struct {
	UUID string
}
