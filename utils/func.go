package utils

import (
	"path/filepath"
	"os"
	"io"
	"archive/tar"
	"fmt"
	"time"
	"bytes"
	"net/http"
	"compress/gzip"
	"log"
	"encoding/json"
	"crypto/md5"
	"encoding/hex"
	"github.com/jlaffaye/ftp"
)

func Mkdir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}


func DeleteFile(path string) error {
	log.Println("Cleaning "+path)
	err := os.Remove(path)

	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}


func PrintDownloadPercent(done chan int64, path string, total int64, progress *float64) {

	var stop bool = false

	for {
		select {
		case <-done:
			stop = true
		default:

			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()

			if size == 0 {
				size = 1
			}

			var percent float64 = float64(size) / float64(total) * 100

			fmt.Printf("%.0f", percent)
			fmt.Println("%")

			if progress != nil {
				*progress = percent
			}
		}

		if stop {
			if progress != nil {
				*progress = 100
			}
			break
		}

		time.Sleep(time.Second)
	}
}

type Manifest struct {
	Files []File
}

type File struct {
	Path string
	Md5  string
}

func DownloadManifest(url string, manifest *Manifest) error {
	headResp, err := http.Head(url)

	if err != nil {
		return err
	}

	defer headResp.Body.Close()
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(manifest)
}

func DownloadFile(conn *ftp.ServerConn, f File, dest string, progress *float64) {

	log.Printf("Downloading file %s from ftp\n", f.Path)

	var path bytes.Buffer
	path.WriteString(dest)
	path.WriteString("/")
	path.WriteString(f.Path)

	start := time.Now()

	out, err := os.Create(path.String())

	if err != nil {
		fmt.Println(path.String())
		panic(err)
	}

	defer out.Close()

	size, err := conn.FileSize(f.Path)

	if err != nil {
		panic(err)
	}

	done := make(chan int64)

	go PrintDownloadPercent(done, path.String(), int64(size), progress)

	resp, err := conn.Retr(f.Path)

	if err != nil {
		panic(err)
	}

	defer resp.Close()

	n, err := io.Copy(out, resp)

	if err != nil {
		panic(err)
	}

	done <- n

	elapsed := time.Since(start)
	log.Printf("Download completed in %s", elapsed)

	go func () {
		hash, _ := hash_file_md5(path.String())
		if hash != f.Md5 {
			fmt.Errorf("File %s has an invalid MD5 !", path)
		}
	}()
}

func Ungzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}


func hash_file_md5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}