package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/0xAX/notificator"
)

type Info struct {
	Colors       []string `json:"colrs"`
	LoadInfo     []string `json:"loadInfo"`
	LoadInfoYday []string `json:"loadInfoYday"`
}

func getInfoAndWarn() {
	// Request the HTML page.
	res, err := http.Get("https://www.taipower.com.tw/d006/loadGraph/loadGraph/data/loadpara.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	bs := string(body)

	bs = strings.Replace(bs, ";", ",", -1)

	re := regexp.MustCompile("var (.*) =")
	bs = re.ReplaceAllString(bs, "\"$1\":")
	re = regexp.MustCompile("//.*")
	bs = re.ReplaceAllString(bs, "")
	re = regexp.MustCompile("<!--.*-->")
	bs = re.ReplaceAllString(bs, "")
	bs = "{" + bs + "\"paceholder\":[]}"

	info := Info{}
	err = json.Unmarshal([]byte(bs), &info)
	if err != nil {
		fmt.Println("error:", err)
	}

	cur, _ := strconv.ParseFloat(strings.Replace(info.LoadInfo[0], ",", "", -1), 64)
	max, _ := strconv.ParseFloat(strings.Replace(info.LoadInfo[2], ",", "", -1), 64)
	percentage := cur / max * 100

	if percentage > 96.0 {
		notify := notificator.New(notificator.Options{
			DefaultIcon: "icon/default.png",
			AppName:     "Taipower notifier",
		})

		warnmsg := fmt.Sprintf("pwer shortage warning! %f%", percentage)

		notify.Push("Taipower", warnmsg, "/home/user/icon.png", notificator.UR_CRITICAL)
	} else {
		fmt.Println("Cur percentage: " + strconv.FormatFloat(percentage, 'f', -1, 64))
	}
}

// from https://gist.github.com/ryanfitz/4191392
func doEvery(d time.Duration, f func()) {
	f()

	for range time.Tick(d) {
		f()
	}
}

func main() {
	doEvery(10*time.Minute, getInfoAndWarn)
}
