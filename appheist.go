package main

import (
	"bufio"
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

func download(resource *url.URL) []byte {
	log.Println("Downloading HTML..." + resource.String())

	client := &http.Client{}

	req, err := http.NewRequest("GET", resource.String(), nil)
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
			log.Printf("StatusCode: %d\n Try: %d. Waiting....", resp.StatusCode, try)
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
	fmt.Println("Downloading Stream..." + resource.String())

	client := &http.Client{}

	req, err := http.NewRequest("GET", resource.String(), nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; U; CPU iPhone OS 10 like Mac OS X; en-us)")

	var resp *http.Response

	try := 1
	for {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("*** Error occured while downloading: " + err.Error())
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

func parseApps(content string) []string {

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

func parseVersions(content string, appname string) []string {

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

			//fmt.Println(location[1])
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

var (
	developer = ""
)

func readIndexFile() map[string]string {
	file, err := os.OpenFile("./files/index", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	lines := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitline := strings.Split(line, ",")
		app := splitline[0]
		link := splitline[1]

		lines[app] = link
	}

	return lines
}

func appendToIndex(app string, link string) {
	file, err := os.OpenFile("./files/index", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString(app + "," + link + "\n")
}

func main() {
	fmt.Println("APK Downloader by wunderwuzzi23")

	developer = "facebook-2"
	downloadlinks := readIndexFile()

	url, err := url.Parse(mirrorApkRoot + developer)
	if err != nil {
		log.Println(err)
	}

	//retrieve the apps of the developer
	appsHTMLContent := string(download(url))
	apps := parseApps(appsHTMLContent)

	apps = []string{}
	apps = append(apps, "flash")

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

			appHTMLVersions := string(download(uploadsURL))
			versions := parseVersions(appHTMLVersions, appName)

			log.Println("*** Retrieving variants")
			for _, version := range versions {

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
					log.Println("Variant: " + downloadString)

					downloadURL, err := url.Parse(downloadString)
					if err != nil {
						log.Printf("Feed URL %s: %v ", downloadString, err)
						continue
					}

					content := string(download(downloadURL))
					//downloadlinks = getDownloadLink(content)

					for _, v := range getDownloadLink(content) {
						downloadlinks[downloadString] = v
					}
				}
			}

			//are there more versions/pages for the app
			patternNext := fmt.Sprintf("/uploads/page/%d/?q=\"%s\">Next ›</a>", page+1, appName)
			if strings.Contains(appHTMLVersions, patternNext) {
				page++
			} else {
				break
			}

		} //end paging

		//downloading all versions for this app
		log.Println("Files to download:")
		for k, link := range downloadlinks {
			log.Printf(" ====> %s (%s)", link, k)

			downloadURL, err := url.Parse("https://www.apkmirror.com" + link)
			if err != nil {
				log.Printf("Erorr downloading %s %v", link, err)
			}
			streamReader := downloadStream(downloadURL)

			dirlocation := fmt.Sprintf("./files/%s/%s/%s", developer, appName, k)
			if os.MkdirAll(dirlocation, os.ModePerm) != nil {
				log.Printf("Error creating directory file %s: %v ", dirlocation, err)
				continue
			}

			temp := strings.Split(k, "/")
			version := temp[5]
			variant := temp[6]

			filelocation := fmt.Sprintf("./files/%s/%s/%s/%s", developer, appName, version, variant)
			file, err := os.Create(filelocation)
			if err != nil {
				log.Printf("Error saving file %s: %v ", filelocation, err)
				continue
			}
			defer file.Close()

			_, err = io.Copy(file, streamReader)
		}
	}
}
