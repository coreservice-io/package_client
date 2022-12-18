package package_client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SHA256(input []byte) []byte {
	hash := sha256.Sum256(input)
	return hash[:]
}

func UnZipTo(zipfile_path string, Folder string, to_delete_zipfile bool) error {

	content, err := os.ReadFile(zipfile_path) // just pass the file name
	if err != nil {
		return err
	}

	// unzip to temp folder
	err = os.MkdirAll(Folder, 0777)
	if err != nil {
		return errors.New("unzipTo os.MkdirAll err :" + err.Error() + " , filePath" + Folder)
	}

	// gzip read
	body := bytes.NewReader(content)
	gr, err := gzip.NewReader(body)
	if err != nil {
		return errors.New("unzipTo gzip read file error:" + err.Error())
	}
	defer gr.Close()
	// tar read
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("unzipTo file error:" + err.Error())
		}

		arr := strings.Split(h.Name, "/")
		nameArr := []string{}
		for _, v := range arr {
			if v != "" {
				nameArr = append(nameArr, v)
			}
		}
		if len(nameArr) <= 1 {
			continue
		}
		name := filepath.Join(nameArr[1:]...)

		filePath := filepath.Join(Folder, name)
		if h.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, 0777)
			if err != nil {
				return errors.New("unzipTo os.MkdirAll err:" + err.Error() + "filePath:" + filePath)
			}
			continue
		}

		file_content, err := io.ReadAll(tr)
		if err != nil {
			return errors.New("unzipTo io.ReadAll err:" + err.Error() + " , filePath:" + filePath)
		}

		err = os.WriteFile(filePath, file_content, 0777)
		if err != nil {
			return errors.New("Error creating :" + filePath + ", err:" + err.Error())
		}

	}

	if to_delete_zipfile {
		os.Remove(zipfile_path)
	}

	return nil
}

func DownloadFile(save_filepath string, download_url string, filehash string) error {

	// Get the data
	resp, err := http.Get(download_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	file_content, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("download_file io.ReadAll err:" + err.Error() + " , download_url:" + download_url)
	}

	// check hash
	m := SHA256(file_content)
	downloadFileHash := hex.EncodeToString(m)
	if downloadFileHash != filehash {
		return errors.New("download_file hash error")
	}

	err = os.WriteFile(save_filepath, file_content, 0777)
	if err != nil {
		return errors.New("download_file Error creating:" + save_filepath + ", err:" + err.Error())
	}

	return nil
}
