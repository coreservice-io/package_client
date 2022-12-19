package package_client

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

const PACKAGE_SERVICE_URL = "http://api.package.coreservice.io:8080"
const DEFAULT_AUTO_UPDATE_INTERVAL_SECS = 3600 * 24 //1 day
const DEFAULT_AUTO_UPDATE_INTERVAL_SECS_MIN = 3     //min auto update interval is 3 secs

type PackageClient struct {
	Package_id                int
	Current_version           string
	Token                     string
	auto_update_interval_secs int64
	sync_interval_secs_remote bool
	last_update_unixtime      int64

	to_stop                  chan struct{}
	auto_update_running      bool
	auto_update_running_lock sync.Mutex
	auto_update_func         func(*PackageClient, *Msg_resp_app_version, error) error
	auto_update_log_callback func(logstr string)
}

// when ticker arrive which mean it is time to callupdate , the to_update_call will be triggered
// in which you put your real update code there
func NewPackageClient(token string, package_id int, current_version string,
	sync_interval_secs_remote bool, to_update_call func(*PackageClient, *Msg_resp_app_version, error) error, auto_update_log_callback func(logstr string)) (*PackageClient, error) {

	if to_update_call == nil {
		return nil, errors.New("to_update_call required")
	}

	//test version api correct
	_, err := GetRemoteAppVersion(token, package_id)
	if err != nil {
		return nil, err
	}
	return &PackageClient{
		Package_id:                package_id,
		Token:                     token,
		Current_version:           current_version,
		auto_update_interval_secs: DEFAULT_AUTO_UPDATE_INTERVAL_SECS,
		auto_update_func:          to_update_call,
		auto_update_log_callback:  auto_update_log_callback,
		last_update_unixtime:      0,
		sync_interval_secs_remote: sync_interval_secs_remote,
		to_stop:                   make(chan struct{}),
	}, nil
}

func (pc *PackageClient) Log(logstr string) *PackageClient {
	if pc.auto_update_log_callback != nil {
		pc.auto_update_log_callback(logstr)
	}
	return pc
}

func (pc *PackageClient) Update() error {
	pc.last_update_unixtime = time.Now().Unix()
	app_v, app_v_err := pc.GetRemoteAppVersion()

	if app_v_err == nil {

		if pc.sync_interval_secs_remote {
			pc.Log("sync remote auto_update_secs to local")
			pc.auto_update_interval_secs = app_v.Update_secs
		} else {
			pc.Log("use local auto_update_secs instead of using remote ")
		}

		if app_v.Version != pc.Current_version {
			pc.Log("remote v:" + app_v.Version + " ,local v:" + pc.Current_version)
			pc.Log("update function to call")
			update_error := pc.auto_update_func(pc, app_v, app_v_err)
			if update_error == nil {
				pc.Log("update function success ,local version updated")
				pc.Current_version = app_v.Version
				return nil
			} else {
				pc.Log("update function failed ,local version won't get updated")
				return update_error
			}
		} else {
			pc.Log("remote version same to local version, remote v:" + app_v.Version)
			return nil
		}
	} else {
		return app_v_err
	}
}

func (pc *PackageClient) SetAutoUpdateInterval(update_interval_secs int64) *PackageClient {
	pc.auto_update_interval_secs = update_interval_secs
	return pc
}

func (pc *PackageClient) StartAutoUpdate() error {
	pc.auto_update_running_lock.Lock()
	defer pc.auto_update_running_lock.Unlock()

	if pc.auto_update_running {
		return errors.New("some other update is running")
	}

	pc.auto_update_running = true

	go func() {

		for {
			select {
			case <-time.After(DEFAULT_AUTO_UPDATE_INTERVAL_SECS_MIN * time.Second):
				if time.Now().Unix()-pc.last_update_unixtime > pc.auto_update_interval_secs {
					pc.Update()
				}
			case <-pc.to_stop:
				return
			}
		}
	}()

	return nil
}

func (pc *PackageClient) StopAutoUpdate() error {
	pc.auto_update_running_lock.Lock()
	defer pc.auto_update_running_lock.Unlock()

	if !pc.auto_update_running {
		return errors.New("no auto-update is running")
	}
	pc.to_stop <- struct{}{}
	pc.auto_update_running = false
	return nil
}

func (pc *PackageClient) DecodeAppDetail(app_v *Msg_resp_app_version, decode_object interface{}) error {

	err := json.Unmarshal([]byte(app_v.Content), decode_object)
	if err != nil {
		return errors.New("json.Unmarshal app version content error:" + err.Error() + " ,app_v.Content:" + app_v.Content)
	}

	return nil
}

func (pc *PackageClient) GetRemoteAppVersion() (*Msg_resp_app_version, error) {
	return GetRemoteAppVersion(pc.Token, pc.Package_id)
}
