package internal

var ProgressState struct {
	Aborted, Cloning, Inquiring, Completed string
} = struct {
	Aborted, Cloning, Inquiring, Completed string
}{
	"Aborted",
	"Cloning",
	"Inquiring",
	"Completed",
}

type ProgressMessage struct {
	State       string
	Name        string
	UUID        string
	CopiedBytes int64
	TotalBytes  int64
}
