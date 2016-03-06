package main

type imageWriterAbortedError string

func (e imageWriterAbortedError) Error() string {
	return string(e)
}
