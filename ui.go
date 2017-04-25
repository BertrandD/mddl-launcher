package main

import (
	"github.com/andlabs/ui"
	"log"
	"time"
	"github.com/bertrandd/launcher/utils"
	"io/ioutil"
	"runtime"
	"fmt"
	"net/http"
	"strings"
	"os/exec"
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

	bNeedUpdate := false
	latestVersion := "Fetching..."
	currentVersion, err := getCurrentVersion()

	ui.Main(func() {
		//name := ui.NewEntry()
		button := ui.NewButton("Download latest version")
		progress:= ui.NewProgressBar()
		box := ui.NewVerticalBox()
		bottom := ui.NewHorizontalBox()
		top := ui.NewVerticalBox()
		top.SetPadded(true)
		manifest := ui.NewLabel("lorem ipsum dolor sit amet") // TODO : get welcome message from web, display changelog...
		top.Append(manifest, true)

		latestVersionLabel := ui.NewLabel("Latest version : "+latestVersion)
		currentVersionLabel := ui.NewLabel("Current version : "+currentVersion)
		versionBox := ui.NewHorizontalBox()
		versionBox.SetPadded(true)
		versionBox.Append(latestVersionLabel, false)
		versionBox.Append(currentVersionLabel, false)

		top.Append(versionBox, false)
		bottom.SetPadded(true)
		bottom.Append(progress, true)
		bottom.Append(button, false)
		box.Append(top, true)
		box.Append(bottom, false)
		box.SetPadded(true)
		window := ui.NewWindow("MiddleWar", 600, 500, false)
		window.SetChild(box)
		button.OnClicked(func(*ui.Button) {
			button.Disable()
			if bNeedUpdate {
				button.SetText("Updating...")
				var percent float64
				go func() {
					Update(os, latestVersion, &percent)
					currentVersion = latestVersion
					err = ioutil.WriteFile(LOCAl_VERSION_FILE, []byte(latestVersion), 0644)
					check(err)
					bNeedUpdate = false
					ui.QueueMain(func() {
						button.SetText("Play")
						button.Enable()
						currentVersionLabel.SetText("Current version :"+currentVersion)
					})
				}()
				go printProgress(progress, &percent)
			} else {
				Launch(os)
			}
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})

		go func() {
			log.Println("Fetching latest version...")
			latestVersion, err = GetLatestVersion()
			log.Printf("Latest version : %s", latestVersion)
			ui.QueueMain(func() {
				latestVersionLabel.SetText("Latest version : " + latestVersion)
			})

			utils.Mkdir(DIST)

			if err != nil {
				err = ioutil.WriteFile(LOCAl_VERSION_FILE, []byte(latestVersion), 0644)
				check(err)
				bNeedUpdate = true
			}

			log.Println("Current version : " + string(currentVersion))
			bNeedUpdate = bNeedUpdate || (string(currentVersion) != latestVersion)

			if !bNeedUpdate {
				ui.QueueMain(func() {
					button.SetText("Play")
					button.Enable()
				})
			}
		}()

		window.Show()
	})

	if err != nil {
		panic(err)
	}
}


func printProgress(progressBar *ui.ProgressBar, percent *float64) {
	for {
		ui.QueueMain(func() {
			progressBar.SetValue(int(*percent))
		})
		if *percent == 100.0 {
			break
		}
		time.Sleep(time.Second/2)
	}

}



func getCurrentVersion() (string, error) {
	localVersion, err := ioutil.ReadFile(LOCAl_VERSION_FILE)
	return string(localVersion), err
}

func GetLatestVersion() (version string, err error) {
	response, err := http.Get(URL_LATEST)
	if err != nil {
		log.Fatal(err)
		return "", err
	} else {
		defer response.Body.Close()
		responseData, err := ioutil.ReadAll(response.Body)
		version = string(responseData)
		return version, err
	}
}

func Update(os, version string, progress *float64) (err error){
	log.Println("Downloading latest version... ")

	var archiveName = fmt.Sprintf(ARCHIVE_NAME_FORMAT, os)

	url := fmt.Sprintf(ARCHIVE_URL, strings.TrimSpace(version), archiveName)
	utils.DownloadFile(url, DIST, progress)
	defer utils.DeleteFile(DIST + archiveName)
	log.Println("Decompressing...")

	err = utils.Ungzip(DIST+archiveName, DIST+"temp.tar")
	defer utils.DeleteFile(DIST + "temp.tar")
	check(err)
	utils.Untar(DIST+"temp.tar", LOCAL_APP_DIRECTORY)
	return err
}

func Launch(os string) {
	cmd := exec.Command(GetExecutable(os))
	err := cmd.Start()
	check(err)
	log.Println("Launching Game")
	go cmd.Wait()
}


func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
