package main

import (
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func main() {
	info, err := host.Info()
	if err != nil {
		return
	}
	fmt.Println(info)

	go func() {
		ticker := time.Tick(time.Second)
		for {
			select {
			case <-ticker:
			}
		}
	}()
	go func() {
		ticker := time.Tick(time.Second)
		for {

			select {
			case <-ticker:
				{
					target := "chrome.exe"

					processes, _ := process.Processes()
					for _, p := range processes {
						name, err := p.Name()
						if err != nil {
							fmt.Println(err)
							break
						}
						if name == target {
							fmt.Println(name)
							cmd := exec.Command("taskkill", "/f", "/t", "/pid", strconv.Itoa(int(p.Pid)))
							output, err := cmd.CombinedOutput()
							fmt.Println(string(GbkToUtf8(output)))
							if err != nil {
								fmt.Println(err)

							}
						}
					}
				}
			}
		}
	}()
	group := sync.WaitGroup{}
	group.Add(1)
	group.Wait()

}
func GbkToUtf8(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return []byte{}
	}
	return d
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
