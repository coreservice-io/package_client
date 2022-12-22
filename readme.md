## how to use


### use-case-1
#### used for single update ,below example to show a api usage 

```go

package main

import (
	"fmt"

	"github.com/coreservice-io/package_client"
)

func main() {
	package_client.StartCacheRefreshJob()
	fmt.Println(package_client.GetAppVersion("", 22, true)) //from cache ,fast ,used for fast api access ,cache will auto-update every 2 minutes
	fmt.Println(package_client.GetAppVersion("", 22, false))//no cache
}

```

### use-case-2
#### used for single update ,below example to show a local resource update process

```go
func StartAutoUpdate(current_version string, sync_remote_update_secs bool, download_folder string, update_success_callback func(), logger func(string), err_logger func(string)) (*package_client.PackageClient, error) {

	pc, pc_err := package_client.NewPackageClient(AUTO_UPDATE_CONFIG_TOKEN, AUTO_UPDATE_CONFIG_PACKAGEID,
		current_version, sync_remote_update_secs, func(pc *package_client.PackageClient, m *package_client.Msg_resp_app_version) error {

			app_detail_s := &package_client.AppDetail_Standard{}
			decode_err := pc.DecodeAppDetail(m, app_detail_s)
			if decode_err != nil {
				return decode_err
			}

			download_err := package_client.DownloadFile(filepath.Join(download_folder, "temp"), app_detail_s.Download_url, app_detail_s.File_hash)
			if download_err != nil {
				return download_err
			}

			unziperr := package_client.UnZipTo(filepath.Join(download_folder, "temp"), download_folder, true)
			if unziperr != nil {
				return unziperr
			}

			update_success_callback()
			return nil

		}, func(logstr string) {
			logger(logstr)
		}, func(logstr string) {
			err_logger(logstr)
		})

	if pc_err != nil {
		return nil, pc_err
	}

	start_err := pc.SetAutoUpdateInterval(AUTO_UPDATE_CONFIG_UPDATE_INTERVAL_SECS).StartAutoUpdate()
	if start_err != nil {
		return nil, start_err
	}

	return pc, nil
}
```