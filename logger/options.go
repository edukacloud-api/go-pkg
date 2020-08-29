package logger

import "time"

type Options struct {
	FileLocation    string        `json:"fileLocation"`
	FileTdrLocation string        `json:"fileTdrLocation"`
	FileMaxAge      time.Duration `json:"fileMaxAge"`
	Stdout          bool          `json:"stdout"`
}
