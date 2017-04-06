package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"github.com/bertrandd/updater/utils"
	"os/exec"
	"fmt"
	"strings"
	"runtime"
)
var ROOT_URL string
var ARCHIVE_URL string = ROOT_URL+"%s/%s"
var URL_LATEST string = ROOT_URL+"latest"

const DIST = "./dist/"
const LOCAL_APP_DIRECTORY = DIST+"app/"
const LOCAl_VERSION_FILE = DIST+"version"
const ARCHIVE_NAME_FORMAT = "MiddleWar-%s-x64-online.tar.gz"
const APP_DIRECTORY_FORMAT = "MiddleWar-%s-x64/"
const APP_EXE_NAME = "MiddleWar"

func GetExecutable(os string) string {
	var suffix string
	if os == "win" {
		suffix=".exe"
	}
	return fmt.Sprintf(LOCAL_APP_DIRECTORY+APP_DIRECTORY_FORMAT+APP_EXE_NAME+suffix,os)
}

func main() {
	const arch = runtime.GOARCH
	os := runtime.GOOS
	if os == "windows" {
		os = "win"
	}
	//os = "win"

	log.Println("OS: " + os)
	log.Println("Arch:" + arch)
	log.Println("Fetching latest version...")
	response, err := http.Get(URL_LATEST)
	check(err)

	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	check(err)
	latestVersion := string(responseData)
	log.Printf("Latest version : %s", string(latestVersion))

	utils.Mkdir(DIST)

	bNeedUpdate := false

	currentVersion, err := ioutil.ReadFile(LOCAl_VERSION_FILE)
	if err != nil {
		err = ioutil.WriteFile(LOCAl_VERSION_FILE, []byte(latestVersion), 0644)
		check(err)
		bNeedUpdate = true
	}

	log.Println("Current version : " + string(currentVersion))
	bNeedUpdate = bNeedUpdate || (string(currentVersion) != latestVersion)

	if bNeedUpdate {
		log.Print("Downloading latest version... ")

		var archiveName = fmt.Sprintf(ARCHIVE_NAME_FORMAT, os)

		url := fmt.Sprintf(ARCHIVE_URL, strings.TrimSpace(latestVersion), archiveName)
		utils.DownloadFile(url, DIST, nil)
		defer utils.DeleteFile(DIST + archiveName)
		log.Println("Decompressing...")

		err = utils.Ungzip(DIST+archiveName, DIST+"temp.tar")
		defer utils.DeleteFile(DIST + "temp.tar")
		check(err)
		utils.Untar(DIST+"temp.tar", LOCAL_APP_DIRECTORY)
	} else {
		log.Println("Already up to date !")
	}

	cmd := exec.Command(GetExecutable(os))
	err = cmd.Start()
	check(err)
	log.Println("Launching MiddleWar")
	go cmd.Wait()
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
