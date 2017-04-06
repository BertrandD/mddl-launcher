package middlewar

import (
	"net/http"
	"io/ioutil"
	"log"
	"github.com/bertrandd/updater/utils"
)


func GetVersion() (version string, err error) {
	response, err := http.Get("http://url/to/archives/latest")
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

func Update(progress *float64) (err error){
	log.Println("Downloading latest version... ")

	url := "http://url/to/archives/v0.10.12/MiddleWar-linux-x64.tar.gz"
	utils.Mkdir("./dist")
	utils.DownloadFile(url, "./dist/", progress)
	//defer deleteFile("./dist/MiddleWar-linux-x64.tar.gz")
	log.Println("Decompressing...")
	err = utils.Ungzip("./dist/MiddleWar-linux-x64.tar.gz", "./dist/Middlewar.tar")
	defer utils.DeleteFile("./dist/Middlewar.tar")
	if err != nil {
		log.Fatal(err)
	} else {
		utils.Untar("./dist/Middlewar", "./dist/MiddleWar")
	}
	return err
}