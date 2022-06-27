package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

/// To search through all the files:
/// find folder -type f | xargs zgrep -i -f pattern.file
///

func download(resource *url.URL) []byte {
	//log.Println("Downloading HTML..." + resource.String())

	client := &http.Client{}

	req, err := http.NewRequest("GET", resource.String(), nil)
	if err != nil {
		log.Println("*** Error occured NewRequest construction: " + err.Error())
	}
	req.Header.Add("User-Agent", _requestHeader)

	var resp *http.Response

	try := 1
	for {
		resp, err = client.Do(req)
		if err != nil {
			log.Println("*** Error occured while downloading (sleeping a bit, then retry): " + err.Error())
			seconds := time.Duration(try*60) * time.Second
			time.Sleep(seconds)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			break
		}

		if resp.StatusCode != 200 {
			//HTTP/1.1 429 Too Many Requests
			log.Printf("StatusCode: %d. Try: %d. Waiting....\n", resp.StatusCode, try)
			seconds := time.Duration(try*60) * time.Second
			time.Sleep(seconds)
			try++
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("*** Error occured while reading content: " + err.Error())
	}

	return body
}

func downloadStream(resource *url.URL) io.Reader {
	//log.Println("Downloading Stream..." + resource.String())

	client := &http.Client{}

	req, err := http.NewRequest("GET", resource.String(), nil)
	if err != nil {
		log.Println("*** Error occured NewRequest construction: " + err.Error())
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; U; CPU iPhone OS 10 like Mac OS X; en-us)")

	var resp *http.Response

	try := 1
	for {
		resp, err = client.Do(req)
		if err != nil {
			log.Println("*** Error occured while downloading (sleeping a bit then retry): " + err.Error())
			seconds := time.Duration(try*60) * time.Second
			time.Sleep(seconds)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			break
		}

		if resp.StatusCode != 200 {
			//HTTP/1.1 429 Too Many Requests
			log.Printf("StatusCode: %d\n Try: %d. Waiting....", resp.StatusCode, try)
			seconds := time.Duration(try*60) * time.Second
			time.Sleep(seconds)
			try++
		}
	}

	return resp.Body
}

func parseApps(developer string, content string) []string {

	appsearchpattern := "<a class=\"fontBlack\" href=\"/apk/" + developer + "/"
	lines := strings.Split(content, "\n")

	apps := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, appsearchpattern) {
			applocation := strings.Split(line, "/")
			apps[applocation[3]] = 1
		}
	}

	returnapps := []string{}
	for k := range apps {
		returnapps = append(returnapps, k)
	}

	return returnapps
}

func parseVersions(content string, developer string, appname string) []string {

	searchpattern := fmt.Sprintf("<a class=\"fontBlack\" href=\"/apk/%s/%s", developer, appname)
	lines := strings.Split(content, "\n")

	data := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, searchpattern) {
			location := strings.Split(line, "/")
			data[location[4]] = 1
		}
	}

	retversions := []string{}
	for k := range data {
		retversions = append(retversions, k)
	}

	return retversions
}

func parseAppPageCount(content string, developer string, appname string) int {

	searchpattern := ">Page 1 of "
	lines := strings.Split(content, "\n")

	pagecount := 1
	var err error
	for _, line := range lines {
		if strings.Contains(line, searchpattern) {
			location := strings.Split(line, searchpattern)
			//fmt.Println(location[1])
			location = strings.Split(location[1], "<")
			//fmt.Println(location[0])
			pagecount, err = strconv.Atoi(location[0])
			if err != nil {
				log.Printf("Error converting %s to integer", location[0])
			}
			break
		}
	}

	return pagecount
}

func parseVariants(content string, appname string, version string) []string {

	//searchpattern := fmt.Sprintf("/apk/%s/%s/%s/", developer, appname, version)
	searchpattern := "-android-apk-download"
	lines := strings.Split(content, "\n")

	data := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, searchpattern) {
			if strings.Contains(line, "<a class=\"accent_color\" href=\"/apk/") {
				location := strings.Split(line, "/")
				data[location[5]] = 1
			}
		}
	}

	retvariants := []string{}
	for k := range data {
		retvariants = append(retvariants, k)
	}

	return retvariants
}

func getDownloadLink(content string) []string {

	//searchpattern := "https://www.apkmirror.com/wp-content/themes/APKMirror/download.php?"
	//searchpattern := "/wp-content/themes/APKMirror/download.php?id="
	searchpattern := "/download/?key="
	lines := strings.Split(content, "\n")

	data := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, searchpattern) {
			location := strings.Split(line, "href")
			location = strings.Split(location[1], "\"")

			log.Println(location[1])
			data[location[1]] = 1
		}
	}

	retdata := []string{}
	for k := range data {
		retdata = append(retdata, k)
	}

	return retdata
}

const (
	mirrorApkRoot  = "https://www.apkmirror.com/apk/"
	_requestHeader = "Mozilla/5.0 (Linux; U; Android 4.2.2; en-us; SM-T217S Build/JDQ39) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Safari/534.30"
)

func readIndexFile() []string {
	file, err := os.OpenFile("./files/index", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func appendToIndex(developer string, appname string, version string, variant string, link string) {
	file, err := os.OpenFile("./files/index", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	line := fmt.Sprintf("%s,%s,%s,%s,%s\n", developer, appname, version, variant, link)
	file.WriteString(line)
}

func downloadFile(developer string, appName string, version string, variant string, link string) {
	//downloading all versions for this app
	log.Printf(" ====> %s\n", link)

	downloadURL, err := url.Parse("https://www.apkmirror.com" + link)
	if err != nil {
		log.Printf("Error downloading %s %v", link, err)
	}

	//streamReader := downloadStream(downloadURL)
	client := &http.Client{}
	req, err := http.NewRequest("GET", downloadURL.String(), nil)
	if err != nil {
		log.Println("*** Error occured NewRequest construction: " + err.Error())
	}
	req.Header.Add("User-Agent", _requestHeader)

	var resp *http.Response

	try := 1
	for {
		resp, err = client.Do(req)
		if err != nil {
			log.Println("*** Error occured while downloading (sleeping a bit, then retry): " + err.Error())
			seconds := time.Duration(try*60) * time.Second
			time.Sleep(seconds)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			break
		}

		if resp.StatusCode != 200 {
			//HTTP/1.1 429 Too Many Requests
			log.Printf("StatusCode: %d. Try: %d. Waiting....", resp.StatusCode, try)
			seconds := time.Duration(try*60) * time.Second
			time.Sleep(seconds)
			try++
		}
	}

	dirlocation := fmt.Sprintf("./files/%s/%s/%s", developer, appName, version)
	if os.MkdirAll(dirlocation, os.ModePerm) != nil {
		log.Printf("Error creating directory file %s: %v ", dirlocation, err)
	}

	filelocation := fmt.Sprintf("./files/%s/%s/%s/%s", developer, appName, version, variant)
	file, err := os.Create(filelocation)
	if err != nil {
		log.Printf("Error saving file %s: %v \n", filelocation, err)
	}
	defer file.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Printf("Error storing file %s.", filelocation)
	}
	log.Printf("%s file stored. Size %d\n", filelocation, size)
}

// var _logfile *os.File
// var _indexcache map[string]string
// var cacheMutex = &sync.Mutex{}

// func reloadCache() {
// 	cacheMutex.Lock()
// 	//load cache/index file
// 	lines := readIndexFile()
// 	_indexcache = make(map[string]string)
// 	for _, line := range lines {
// 		_indexcache[line] = "exists"
// 	}

// 	cacheMutex.Unlock()
// }

// func updateCacheEntries() {
// 	cacheMutex.Lock()
// 	for _, v := range _indexcache {
// 		if _indexcache[v] == "new" {
// 			items := strings.Split(v, ",")
// 			appendToIndex(items[0], items[1], items[2], items[3], items[4])
// 		}
// 	}
// 	cacheMutex.Unlock()
// }

// //this will keep running every 60 seconds
// func refreshCache() {
// 	log.Println("*** Updating cache")
// 	updateCacheEntries()
// 	reloadCache()
// 	time.AfterFunc(60*time.Second, refreshCache)
// }

func main() {
	log.Println("AppHeist Downloader by wunderwuzzi23")

	//logfile
	starttime := time.Now()
	os.Mkdir("logs", 0744)
	filename := "./logs/appheist." + starttime.Format("2006-01-02_150405") + ".log"
	_logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	multiwriter := io.MultiWriter(os.Stdout, _logfile)
	log.SetOutput(multiwriter)

	//flags
	var developer string
	var app string
	var mode string
	var pagingStart int
	var skipVariants bool
	flag.StringVar(&developer, "developer", "", "Developer account")
	flag.StringVar(&app, "app", "", "Which app?")
	flag.BoolVar(&skipVariants, "skipvariants", true, "Only index first found variant for version")
	flag.IntVar(&pagingStart, "pagestart", 1, "Specify the page to start enumerating")
	flag.StringVar(&mode, "mode", "buildindex", "buildindex, download, listapps, listapps+")

	flag.Parse()

	log.Printf("Options")
	log.Printf("Dev: %s App: %s Mode: %s PagingStart: %d SkipVarints: %v", developer, app, mode, pagingStart, skipVariants)

	//construct main url
	url, err := url.Parse("https://www.apkmirror.com/?post_type=app_release&searchtype=apk&s=" + developer)
	if err != nil {
		log.Println(err)
	}

	if mode == "listapps" {
		log.Printf("Use listapps+ for page count for each app (indicator of number of downloads)\n")

		if app == "" {
			appsHTMLContent := string(download(url))
			apps := parseApps(developer, appsHTMLContent)
			for _, appName := range apps {
				log.Printf("App: %s\n", appName)
			}
		} else {
			log.Printf("Can't specify an app with -listapps.")
		}

		return
	}

	if mode == "listapps+" {
		if app == "" {
			appsHTMLContent := string(download(url))
			apps := parseApps(developer, appsHTMLContent)
			for _, appName := range apps {

				//retrieve number of pages for each app
				appuploadsString := fmt.Sprintf("https://www.apkmirror.com/uploads/page/1/?q=%s", appName)
				uploadsURL, err := url.Parse(appuploadsString)
				if err != nil {
					log.Printf("AppUpload URL %s: %v ", appuploadsString, err)
					continue
				}

				appHTMLVersions := string(download(uploadsURL))
				pagecount := parseAppPageCount(appHTMLVersions, developer, appName)

				log.Printf("App: %-35s  PageCount: %d\n", appName, pagecount)
			}
		} else {
			log.Printf("Can't specify an app with -listapps.")
		}

		return
	}

	if mode == "download" {
		lines := readIndexFile()
		numLines := len(lines)
		log.Printf("Files to download: %d", numLines)
		for i, v := range lines {
			log.Printf("Processing line (%d/%d): %s", i, numLines, v)
			line := strings.Split(v, ",")

			//check if file already downloaded?
			checkPath := fmt.Sprintf("./files/%s/%s/%s/%s", line[0], line[1], line[2], line[3])
			_, err := os.Stat(checkPath)
			if err == nil {
				log.Printf("*** Skipping: File %s already found. Skipping download.\n", line[3])
				continue
			}

			downloadFile(line[0], line[1], line[2], line[3], line[4])

			//be nice to the API
			rand.Seed(time.Now().UnixNano())
			n := time.Duration(rand.Intn(180) + 60)
			n = 1
			time.Sleep(n * time.Second)
		}

		log.Printf("Done.")
		return
	}

	//refreshCache()
	lines := readIndexFile()
	fmt.Println("Index loaded. Continuing...")

	//retrieve the apps of the developer
	var apps = []string{app}
	//overwrite if no app filter specified
	if app == "" {
		appsHTMLContent := string(download(url))
		apps = parseApps(developer, appsHTMLContent)
	}

	log.Println("*** Retrieving apps")
	for _, appName := range apps {

		page := pagingStart
		for { //paging

			//https://www.apkmirror.com/uploads/page/1/?q=facebook
			appuploadsString := fmt.Sprintf("https://www.apkmirror.com/uploads/page/%d/?q=%s", page, appName)
			uploadsURL, err := url.Parse(appuploadsString)
			if err != nil {
				log.Printf("AppUpload URL %s: %v ", appuploadsString, err)
				continue
			}

			time.Sleep(1 * time.Second)
			appHTMLVersions := string(download(uploadsURL))
			versions := parseVersions(appHTMLVersions, developer, appName)

			log.Println("*** Retrieving variants")
			for _, version := range versions {
				//log.Println(version)
				downloadURLHelper := fmt.Sprintf("%s%s/%s/%s", mirrorApkRoot, developer, appName, version)
				downloadURL, err := url.Parse(downloadURLHelper)
				if err != nil {
					log.Printf("Feed URL %s: %v ", downloadURL, err)
					continue
				}

				variantsHTMLContent := string(download(downloadURL))
				variants := parseVariants(variantsHTMLContent, appName, version)

				log.Println("*** Retrieving download links")
				for _, variant := range variants {

					downloadString := fmt.Sprintf("%s%s/%s/%s/%s", mirrorApkRoot, developer, appName, version, variant)
					log.Println("Variant: " + downloadString)

					//quick check to save some time in case we already have the file
					skip := false
					for _, line := range lines {
						entry := fmt.Sprintf("%s,%s,%s,%s,", developer, appName, version, variant)
						if strings.Contains(line, entry) {
							log.Printf("Entry %s already found in index. Skipping", entry)
							skip = true
							break
						}
					}

					if skip {
						continue
					}

					downloadURL, err := url.Parse(downloadString)
					if err != nil {
						log.Printf("Feed URL %s: %v ", downloadString, err)
						continue
					}
					content := string(download(downloadURL))
					//downloadlinks = getDownloadLink(content)

					for _, v := range getDownloadLink(content) {
						//to do, chnage to a background routine and use caching
						appendToIndex(developer, appName, version, variant, v)
					}

					log.Printf("Retrieved %s\n", downloadString)

					//if user choose to only download one variant per version exit here
					if skipVariants {
						fmt.Printf("Skipping Variants...\n")
						break
					}
				}
			}

			//are there more versions/pages for the app
			patternNext := fmt.Sprintf("/uploads/page/%d/?q=%s\">Next ›</a>", page+1, appName)
			if strings.Contains(appHTMLVersions, patternNext) {
				page++
				log.Printf("Performing paging: %d", page)
			} else {
				log.Printf("No more pages: %d", page)
				break
			}

		} //end paging
	}
}
