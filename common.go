package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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

func CreatDirIfNotHad(_dir string){
	exist, err := PathExists(_dir)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return
	}

	if exist {
		fmt.Printf("has dir![%v]\n", _dir)
	} else {
		fmt.Printf("no dir![%v]\n", _dir)
		// 创建文件夹
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		} else {
			fmt.Printf("mkdir success!\n")
		}
	}
}

func splitArray(arr []string, num int) ([][]string) {
	max := len(arr)
	if max < num {
		return nil
	}
	var segmens =make([][]string,0)
	quantity:=max/num
	end := 0
	for i := 1; i <= num; i++ {
		qu:=i*quantity
		if i != num {
			segmens= append(segmens,arr[i-1+end:qu])
		}else {
			segmens= append(segmens,arr[i-1+end:])
		}
		end=qu-i
	}
	return segmens
}

func HttpGet(url string)(result string,err error){
	resp ,err1 :=http.Get(url)
	if err1!=nil{
		err=err1
		return
	}
	defer resp.Body.Close()
	buf := make([]byte,4096)
	for {
		n,err2 :=resp.Body.Read(buf)
		if n==0{
			break
		}
		if err2!=nil&&err2!=io.EOF{
			err =err2
			return
		}
		result +=string(buf[:n])

	}
	return
}