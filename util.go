package package_client

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func SHA256(input []byte) []byte {
	hash := sha256.Sum256(input)
	return hash[:]
}

func UnZipTo(zipfile_path string, Folder string, to_delete_zipfile bool) error {

	f, err := os.Open(zipfile_path)
	if err != nil {
		return err
	}
	defer f.Close()

	// unzip to temp folder
	err = os.MkdirAll(Folder, 0755)
	if err != nil {
		return fmt.Errorf("unzipTo os.MkdirAll err:%s, filePath:%s", err.Error(), Folder)
	}

	// gzip read
	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("unzipTo gzip read file err:%s", err.Error())
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
			return fmt.Errorf("unzipTo file err:%s", err.Error())
		}

		filePath := filepath.Join(Folder, h.Name)
		if h.FileInfo().IsDir() {
			if err = os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("unzipTo os.MkdirAll err:%s filePath:%s", err.Error(), filePath)
			}
			continue
		}

		fw, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(h.Mode))
		if err != nil {
			return fmt.Errorf("Error creating:%s, err:%s", filePath, err.Error())
		}
		if _, err := io.Copy(fw, tr); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		fw.Close()
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
		return fmt.Errorf("download_file io.ReadAll err:%s , download_url:%s", err.Error(), download_url)
	}

	// check hash
	m := SHA256(file_content)
	downloadFileHash := hex.EncodeToString(m)
	if downloadFileHash != filehash {
		return errors.New("download_file hash error")
	}

	err = os.WriteFile(save_filepath, file_content, 0777)
	if err != nil {
		return fmt.Errorf("download_file Error creating:%s, err:%s", save_filepath, err.Error())
	}

	return nil
}

type Version struct {
	Head int
	Mid  int
	Tail int
}

func ParseVersion(v string) (*Version, error) {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)
	v = strings.TrimPrefix(v, "v")

	aVersion := strings.Split(v, ".")
	if len(aVersion) != 3 {
		return nil, errors.New("version format error")
	}

	result := &Version{}

	head, err_head := strconv.Atoi(aVersion[0])
	if err_head != nil {
		return nil, err_head
	} else {
		result.Head = head
	}

	mid, err_mid := strconv.Atoi(aVersion[1])
	if err_mid != nil {
		return nil, err_mid
	} else {
		result.Mid = mid
	}

	tail, err_tail := strconv.Atoi(aVersion[2])
	if err_tail != nil {
		return nil, err_tail
	} else {
		result.Tail = tail
	}

	return result, nil
}

func version_num_compare(a_int int, b_int int) int {
	if a_int > b_int {
		return 1
	} else if a_int < b_int {
		return -1
	} else {
		return 0
	}
}

// VersionCompare
// compare version (x.x.x)
// a>b return 1
// a==b return 0
// a<b  return -1
func VersionCompare(a, b string) (int, error) {
	a_v, a_err := ParseVersion(a)
	if a_err != nil {
		return 0, a_err
	}

	b_v, b_err := ParseVersion(b)
	if b_err != nil {
		return 0, b_err
	}

	if c_h := version_num_compare(a_v.Head, b_v.Head); c_h == 0 {
		if c_m := version_num_compare(a_v.Mid, b_v.Mid); c_m == 0 {
			return version_num_compare(a_v.Tail, b_v.Tail), nil
		} else {
			return c_m, nil
		}
	} else {
		return c_h, nil
	}
}
