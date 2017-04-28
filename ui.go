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
	"net/http"
)

var FTP_URL string
var VERSION_URL string
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
	CurrentFile utils.File
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
		box := ui.NewVerticalBox()
			top := ui.NewVerticalBox()
				top.SetPadded(true)
				manifest := ui.NewLabel("lorem ipsum dolor sit amet") // TODO : get welcome message from web, display changelog...
				top.Append(manifest, true)
				versionBox := ui.NewHorizontalBox()
					latestVersionLabel := ui.NewLabel("Latest version : "+latestVersion)
					currentVersionLabel := ui.NewLabel("Current version : "+currentVersion)
					versionBox.SetPadded(true)
					versionBox.Append(latestVersionLabel, false)
					versionBox.Append(currentVersionLabel, false)
				top.Append(versionBox, false)
			bottom := ui.NewHorizontalBox()
				bottom.SetPadded(true)
				progressBar := ui.NewProgressBar()
				button := ui.NewButton("Download latest version")
				bottom.Append(progressBar, true)
				bottom.Append(button, false)
			fetching := ui.NewLabel("")

			box.Append(top, true)
			box.Append(bottom, false)
			box.Append(fetching, false)
			box.SetPadded(true)

		window := ui.NewWindow("MiddleWar", 600, 500, false)
		window.SetChild(box)
		button.OnClicked(func(*ui.Button) {
			button.Disable()
			if bNeedUpdate {
				button.SetText("Updating...")
				progress := Progress{}
				go func() {
					Update(os, &progress, progressBar, fetching)
					currentVersion, _ = getCurrentVersion()
					bNeedUpdate = false
					ui.QueueMain(func() {
						button.SetText("Play")
						button.Enable()
						currentVersionLabel.SetText("Current version : "+currentVersion)
					})
				}()
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

func getCurrentVersion() (string, error) {
	localVersion, err := ioutil.ReadFile(LOCAl_VERSION_FILE)
	return string(localVersion), err
}

func GetLatestVersion() (version string, err error) {
	response, err := http.Get(VERSION_URL)
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

func Update(OS string, progress *Progress, progressBar *ui.ProgressBar, fetching *ui.Label) (err error){
	log.Println("Downloading latest version... ")

	log.Printf("Connecting to %s...\n", FTP_URL)
	client, err := goftp.Dial(FTP_URL)
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


	localManifest, err := utils.GetLocalManifest(LOCAL_APP_DIRECTORY)
	check(err)

	progress.Current = int64(0)
	progress.Total = manifest.Size
	files := []utils.File{}
	for _, file := range manifest.Files {
		if !hasGoodMd5(localManifest, file) {
			files = append(files, file)
		}
	}

	for _, file := range files {
		utils.DownloadFile(client, file, FILES_DIR, LOCAL_APP_DIRECTORY)
		progress.CurrentFile = file
		progress.Current += file.Size
		progress.Percent = float64(progress.Current) / float64(manifest.Size) * 100
		ui.QueueMain(func() {
			fetching.SetText("Fetching file "+progress.CurrentFile.Path)
			progressBar.SetValue(int(progress.Percent))
		})
	}

	ui.QueueMain(func() {
		fetching.SetText("")
		progressBar.SetValue(int(100))
	})

	return err
}

func hasGoodMd5(manifest *utils.Manifest, file utils.File) bool {
	for _, v := range manifest.Files {
		if "./"+v.Path == LOCAL_APP_DIRECTORY+file.Path {
			return v.Md5 == file.Md5
		}
	}
	return false
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

