package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func main() {
	var start, end int

	fmt.Println("Please input start page number:")
	fmt.Scan(&start)

	fmt.Println("Please input end page number:")
	fmt.Scan(&end)

	//Start to process
	doWork(start, end)

}

func doWork(start int, end int) {
	channel := make(chan int)
	for i := start; i <= end; i++ {
		go crawlPic(i, channel)
	}

	for i := start; i <= end; i++ {
		fmt.Printf("page %d has completed!", <-channel)
	}

}

func crawlPic(page int, channel chan int) {
	url := "https://www.douyu.com/wgapi/ordnc/live/web/room/yzList/" + strconv.Itoa(page)
	fmt.Println("url:", url)
	result, err := HttpGet(url)
	if err != nil {
		fmt.Println("HttpGet err:", err)
		return
	}
	//title
	titleRegex := regexp.MustCompile("\"nn\":\"(.*?)\"")
	titles := titleRegex.FindAllStringSubmatch(result, -1)

	//pic
	picRegex := regexp.MustCompile("\"rs16_avif\":\"(.*?)\"")
	picUrls := picRegex.FindAllStringSubmatch(result, -1)

	picMap := make(map[string]string, 240)

	for i := 0; i < len(titles); i++ {
		//store the pic name and pic url as key,value
		picMap[titles[i][1]] = picUrls[i][1]
	}
	//download the pic to local storage
	for title, picUrl := range picMap {

		//create file
		fileName := title + ".png"
		file, err := os.Create("/Users/evan/Downloads/" + fileName)
		if err != nil {
			fmt.Println("os.Create err:", err)
			return
		}

		if picUrl == "" {
			fmt.Println("Skip this picture========>title:" + title + " is invalid , picUrl is :" + picUrl)
			continue
		}
		downloadErr := DownloadFile(picUrl, file)
		if downloadErr != nil {
			fmt.Println("DownloadFile err:", err)
			return
		}
		file.Close()
	}
	channel <- page

}

func DownloadFile(url string, destFile *os.File) (err error) {
	fmt.Println("url:", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "myClient")
	resp, httpErr := client.Do(req)
	if err != nil {
		fmt.Println("http.NewRequest err:", err)
		err = httpErr
		return err
	}
	defer resp.Body.Close()
	buf := make([]byte, 4098)

	for {
		n, readErr := resp.Body.Read(buf)

		if n == 0 {
			break
		}

		if readErr != nil && readErr != io.EOF {
			fmt.Println("resp.bod.Read err:", err)
			continue
		}

		destFile.Write(buf[:n])
	}

	return err
}

func HttpGet(url string) (result string, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	//add user-agent to simulate browser's behaviour
	req.Header.Add("User-Agent", "myClient")
	resp, httpErr := client.Do(req)
	if httpErr != nil {
		fmt.Println("http.Get err:", err)
		err = httpErr
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 4098)
	for {
		//read response body stream
		n, readErr := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		if readErr != nil && readErr != io.EOF {
			fmt.Println("resp.Body.Read err:", readErr)
			err = readErr
			break
		}
		result += string(buf[:n])
		//fmt.Println("result:", result)
	}
	return result, err
}
