package internal

type ProgressMessage struct {
	Name        string
	UUID        string
	CopiedBytes int64
	TotalBytes  int64
	Count       int64
}
