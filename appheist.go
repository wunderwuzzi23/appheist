package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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
	//req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone 10); en-us)")
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; U; CPU iPhone OS 10 like Mac OS X; en-us)")
	//req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	var resp *http.Response

	try := 5
	for {
		resp, err = client.Do(req)
		if err != nil {
			log.Println("*** Error occured while downloading: " + err.Error())
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

	try := 5
	for {
		resp, err = client.Do(req)
		if err != nil {
			log.Println("*** Error occured while downloading: " + err.Error())
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

	appsearchpattern := "<a class=\"downloadLink\" href=\"/apk/" + developer + "/"
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

	searchpattern := fmt.Sprintf("<a class=\"downloadLink\" href=\"/apk/%s/%s", developer, appname)
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

func parseVariants(content string, appname string, version string) []string {

	//searchpattern := fmt.Sprintf("/apk/%s/%s/%s/", developer, appname, version)
	searchpattern := "-android-apk-download"
	lines := strings.Split(content, "\n")

	data := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, searchpattern) {
			location := strings.Split(line, "/")
			data[location[5]] = 1
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
	searchpattern := "/wp-content/themes/APKMirror/download.php?id="
	lines := strings.Split(content, "\n")

	data := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, searchpattern) {
			location := strings.Split(line, "href")
			location = strings.Split(location[1], "\"")

			//log.Println(location[1])
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
	mirrorApkRoot = "https://www.apkmirror.com/apk/"
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
	file, err := os.OpenFile("./files/index", os.O_CREATE|os.O_APPEND, 0755)
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
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; U; CPU iPhone OS 10 like Mac OS X; en-us)")

	var resp *http.Response

	try := 1
	for {
		resp, err = client.Do(req)
		if err != nil {
			log.Println("*** Error occured while downloading: " + err.Error())
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

var _logfile *os.File

func main() {
	log.Println("APK Downloader by wunderwuzzi23")

	//logfile
	starttime := time.Now()
	os.Mkdir("logs", 0744)
	filename := "./logs/aptheist." + starttime.Format("2006-01-02_150405") + ".log"
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
	flag.StringVar(&developer, "developer", "facebook-2", "Developer account")
	flag.StringVar(&app, "app", "", "Which app?")
	flag.StringVar(&mode, "mode", "buildindex", "buildindex or download")

	flag.Parse()

	log.Printf("Options")
	log.Printf("Dev: %s App: %s Mode: %s", developer, app, mode)

	if mode == "download" {
		lines := readIndexFile()
		for _, v := range lines {
			log.Printf("Processing line: " + v)
			line := strings.Split(v, ",")
			downloadFile(line[0], line[1], line[2], line[3], line[4])
			time.Sleep(2 * time.Second)
		}

		log.Printf("Done.")
		return
	}

	url, err := url.Parse(mirrorApkRoot + developer)
	if err != nil {
		log.Println(err)
	}

	//retrieve the apps of the developer
	var apps = []string{app}
	//overwrite if no app filter specified
	if app == "" {
		appsHTMLContent := string(download(url))
		apps = parseApps(developer, appsHTMLContent)
	}

	log.Println("*** Retrieving apps")
	for _, appName := range apps {

		page := 1
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
					downloadString := fmt.Sprintf("%s%s/%s/%s/%s/download", mirrorApkRoot, developer, appName, version, variant)
					//log.Println("Variant: " + downloadString)

					//todo: caching and check if files index contains entry already...

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
