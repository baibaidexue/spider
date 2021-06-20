package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

func CreatDirIfNotHad(_dir string) {
	exist, err := PathExists(_dir)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return
	}

	if exist == false {
		fmt.Printf("[%v] directory not exists.\n", _dir)
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			fmt.Printf("[%v] directory create failed:%v!\n", _dir, err)
		} else {
			fmt.Printf("[%v] directory created success!\n", _dir)
		}
	}
}

// 多个goroutine各自负责一片区域逻辑变更为总体按照顺序下载
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
	resp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n, err2 := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return
		}
		result += string(buf[:n])

	}
	return
}

func openExplorerFolder(path string) {
	curPath := GetCurrentPath()
	finalPath := curPath + "\\" + path
	cmd := exec.Command("explorer.exe", finalPath)
	cmd.Run()
}

//func getCurrentAbPathByCaller() string {
//	var abPath string
//	_, filename, _, ok := runtime.Caller(0)
//	if ok {
//		abPath = path.Dir(filename)
//	}
//	return abPath
//}

func GetCurrentPath() string {
	var projectPath string
	projectPath, _ = os.Getwd()
	return projectPath
}
