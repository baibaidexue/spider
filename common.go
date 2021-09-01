package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/lxn/walk"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreatDirIfNotHad(_dir string) error {
	exist, err := PathExists(_dir)
	if err != nil {
		return err
	}

	if exist == false {
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func splitArray(arr []string, num int) [][]string {
	max := len(arr)
	if max < num {
		return nil
	}
	var segments = make([][]string, 0)
	for i := 0; i < num; i++ {
		segments = append(segments, nil)
	}
	for i := 0; i < max; {
		for j := 0; j < num; j++ {
			segments[j] = append(segments[j], arr[i])
			i++
			if i == max {
				return segments
			}
		}
	}
	return segments
}

func HttpGet(url string) (result string, err error) {
	client := http.Client{
		Timeout: time.Duration(20 * time.Second),
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func openExplorerFolder(path string) {
	curPath := GetCurrentPath()
	finalPath := curPath + "\\" + path
	cmd := exec.Command("explorer.exe", finalPath)
	cmd.Run()
}

func GetCurrentPath() string {
	var projectPath string
	projectPath, _ = os.Getwd()
	return projectPath
}

func NotifyBarCall(title, body string, form walk.Form) {
	ni, err := walk.NewNotifyIcon(form)
	if err != nil {
		return
	}
	defer ni.Dispose()
	ni.SetVisible(true)
	ni.ShowInfo(title, body)
}

func timeCost(start time.Time, funcname string) {
	tc := time.Since(start)
	log.Printf("%v #### Time cost = %v\n", funcname, tc)
}

func runFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func fileCopy(sourceFile, destinationFile string) {
	source, err := os.Open(sourceFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer source.Close()

	destination, err := os.Create(destinationFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		log.Println(err)
	}
}
