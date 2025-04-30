package ffmpeg

import (
	"embed"
	"io/fs"
)

//go:embed bin/win32.zip
var ffmpeg embed.FS

const bin = "bin/win32.zip"

func open() (fs.File, error) {
	return ffmpeg.Open(bin)
}
