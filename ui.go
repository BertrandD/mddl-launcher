package main

import (
	"log"
	"github.com/bertrandd/launcher/utils"
	"runtime"
	"fmt"
	"encoding/json"
	"github.com/secsy/goftp"
	"io/ioutil"
	"os/exec"
	"github.com/andlabs/ui"
	"time"
	"net/http"
)

var ROOT_URL string
var URL_LATEST string = ROOT_URL+"latest"
var MANIFEST_NAME string = "manifest.json"

const LOCAL_APP_DIRECTORY = "./MiddleWar/"
const LOCAl_VERSION_FILE = LOCAL_APP_DIRECTORY+"version"
const APP_EXE_NAME = "MiddleWar"

func GetExecutable(os string) string {
	var suffix string
	if os == "win" {
		suffix=".exe"
	}
	return fmt.Sprintf(LOCAL_APP_DIRECTORY+APP_EXE_NAME+suffix)
}

type Progress struct {
	Total int64
	Current int64
	Percent float64
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
				prog := Progress{}
				go func() {
					Update(os, &prog)
					currentVersion = latestVersion
					err = ioutil.WriteFile(LOCAl_VERSION_FILE, []byte(latestVersion), 0644)
					check(err)
					bNeedUpdate = false
					ui.QueueMain(func() {
						button.SetText("Play")
						button.Enable()
						currentVersionLabel.SetText("Current version : "+currentVersion)
					})
				}()
				go printProgress(progress, &prog)
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

			utils.Mkdir(LOCAL_APP_DIRECTORY)

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


func printProgress(progressBar *ui.ProgressBar, progress *Progress) {
	for {
		ui.QueueMain(func() {
			progressBar.SetValue(int(progress.Percent))
		})
		if progress.Percent == 100.0 {
			break
		}
		time.Sleep(time.Second/8)
	}
}



func getCurrentVersion() (string, error) {
	localVersion, err := ioutil.ReadFile(LOCAl_VERSION_FILE)
	return string(localVersion), err
}

func GetLatestVersion() (version string, err error) {
	return "v0.10.25", nil
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

func Update(OS string, progress *Progress) (err error){
	log.Println("Downloading latest version... ")

	log.Printf("Connecting to %s...\n", ROOT_URL)
	client, err := goftp.Dial(ROOT_URL)
	if err != nil {
		panic(err)
	}

	FILES_DIR := OS+"-x64/"

	log.Println("Retrieving manifest..."+FILES_DIR+MANIFEST_NAME)
	utils.DownloadFile(client, utils.File{Path:MANIFEST_NAME, Mode:420}, FILES_DIR, LOCAL_APP_DIRECTORY)

	manifest := utils.Manifest{}

	file, err := ioutil.ReadFile(LOCAL_APP_DIRECTORY+MANIFEST_NAME)

	err = json.Unmarshal(file, &manifest)
	check(err)

	progress.Current = int64(0)
	progress.Total = manifest.Size
	for _, file := range manifest.Files {
		utils.DownloadFile(client, file, FILES_DIR, LOCAL_APP_DIRECTORY)
		progress.Current += file.Size
		progress.Percent = float64(progress.Current) / float64(manifest.Size) * 100
	}

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
