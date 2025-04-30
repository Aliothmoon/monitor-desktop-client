package ffmpeg

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const Ffmpeg = "lib/ffmpeg/ffmpeg.exe"

func UnPack() error {
	_, err := os.Stat(Ffmpeg)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		file, err := open()
		if err != nil {
			log.Println(err)
			return err
		}
		defer file.Close()
		err = os.MkdirAll("lib/ffmpeg", 0777)
		if err != nil {
			log.Println(err)
			return err
		}
		temp, err := os.CreateTemp(os.TempDir(), "ffmpeg")
		if err != nil {
			log.Println(err)
			return err
		}
		abs, _ := filepath.Abs(temp.Name())
		fmt.Println(abs)
		defer func() {
			go func() {
				_ = temp.Close()
				if err = os.Remove(abs); err != nil {
					log.Println(err)
				}
			}()
		}()
		_, err = io.Copy(temp, file)
		if err != nil {
			log.Println(err)
			return err
		}
		reader, err := zip.OpenReader(abs)

		if err != nil {
			return err
		}
		defer reader.Close()

		exe, err := reader.Open("ffmpeg.exe")
		if err != nil {
			return err
		}
		defer exe.Close()

		target, err := os.OpenFile(Ffmpeg, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Println(err)
			return err
		}
		defer target.Close()
		_, err = io.Copy(target, exe)
		return err
	}
	log.Println(err)
	return err

}

func Version() error {
	f, _ := filepath.Abs(Ffmpeg)
	cmd := exec.Command(f, "-version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

const scale = "scale=1280:720"

func RtmpPushScreen(url string) error {
	f, _ := filepath.Abs(Ffmpeg)
	cmd := exec.Command(f,
		"-f", "gdigrab",
		"-framerate", "30",
		"-i", "desktop",
		"-vf", scale,
		"-c:v", "libx264",
		"-tune", "zerolatency",
		"-crf", "23",
		"-b:v", "2000k",
		"-g", "60",
		"-pix_fmt",
		"yuv420p", "-an",
		"-f", "flv",
		"-vsync", "passthrough",
		url,
	)

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	// DEBUG
	//cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		log.Println("Start Process Error", err)
		return err
	}
	log.Println("Start Process")
	return cmd.Wait()
}
