package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	//page number
	var start, end int

	fmt.Println("Please input start page number:")
	fmt.Scan(&start)
	fmt.Println("Please input end page number:")
	fmt.Scan(&end)

	doWork(start, end)
}

func doWork(start int, end int) {
	channel := make(chan int, end-start+1)
	//crawl data by page
	for i := start; i <= end; i++ {
		go crawlDB(i, channel)

	}
	//monitor the task status
	for i := start; i <= end; i++ {
		fmt.Printf("Pave %d has completed\n", <-channel)
	}
}

func crawlDB(page int, channel chan int) {

	//call douban website
	url := "https://movie.douban.com/top250?start=" + strconv.Itoa((page-1)*25) + "&filter="
	time.Sleep(1 * time.Second)
	result, err := HttpGet(url)
	if err != nil {
		fmt.Println("HttpGet err:", err)
		return
	}

	//Create files
	fileName := "No " + strconv.Itoa(page) + " page.txt"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("os.Create err:", err)
		return
	}

	//Movie Name
	titleRegex := regexp.MustCompile("<img width=\"100\" alt=\"(.*?)\" src=\"(.*?)\" class=\"\">")
	titles := titleRegex.FindAllStringSubmatch(result, -1)
	//Movie Score
	scoreRegex := regexp.MustCompile("<span class=\"rating_num\" property=\"v:average\">(.*?)</span>")
	scores := scoreRegex.FindAllStringSubmatch(result, -1)
	//People count
	pplCountRegex := regexp.MustCompile("<span>(.*?)人评价</span>")
	pplCount := pplCountRegex.FindAllStringSubmatch(result, -1)

	file.WriteString("Tile|Score|pplCount")
	fmt.Println("title count:", len(titles))
	for i := 0; i < len(titles); i++ {
		file.WriteString(titles[i][1] + "|" + scores[i][1] + "|" + pplCount[i][1] + "\n")
	}
	file.Close()
	//Output the page number which has completed
	channel <- page

}

func HttpGet(url string) (result string, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	//add user-agent to simulate browser's behaviour
	req.Header.Add("User-Agent", "myClient")
	resp, httpErr := client.Do(req)

	fmt.Println("resp:", resp)
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
		fmt.Println("result:", result)
	}
	return result, err
}
