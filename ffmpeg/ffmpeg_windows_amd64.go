package ffmpeg

import (
	"embed"
	"io/fs"
)

//go:embed bin/win64.zip
var ffmpeg embed.FS

const bin = "bin/win64.zip"

func open() (fs.File, error) {
	return ffmpeg.Open(bin)
}
