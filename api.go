package package_client

import (
	"encoding/json"
	"errors"
	"strconv"
)

//api meta

type API_META_VERSION struct {
	Meta_version int `json:"meta_version"`
}

// assign version
func (apim *API_META_VERSION) MetaVersion(version int) {
	apim.Meta_version = version
}

type API_META_STATUS struct {
	Meta_status  int    `json:"meta_status"`
	Meta_message string `json:"meta_message"`
}

// assign status
func (apim *API_META_STATUS) MetaStatus(status int, message string) {
	apim.Meta_message = message
	apim.Meta_status = status
}

// @Description Msg_resp_app
type Msg_resp_app_version struct {
	API_META_STATUS
	Version               string `json:"version"`
	Content               string `json:"content"`
	Update_secs           int64  `json:"update_secs"`
	Minimum_allow_version string `json:"minimum_allow_version"`
}

type AppDetail_Standard struct {
	Download_url string `json:"download_url"`
	File_hash    string `json:"file_hash"`
	Exe_name     string `json:"exe_name"`
	Compatible   string `json:"compatible"`
}

// decode_pointer is the adress of your unmarshal object
func GetRemoteAppDetail(token string, package_id int, decode_object interface{}) error {

	v_result, v_err := GetRemoteAppVersion(token, package_id)
	if v_err != nil {
		return v_err
	}

	err := json.Unmarshal([]byte(v_result.Content), decode_object)
	if err != nil {
		return errors.New("json.Unmarshal app version content error:" + err.Error() + " , package_id:" + strconv.Itoa(package_id))
	}

	return nil
}

func GetRemoteAppVersion(token string, package_id int) (*Msg_resp_app_version, error) {

	// request app version info
	request_url := PACKAGE_SERVICE_URL + "/api/version/"
	request_url = request_url + strconv.Itoa(package_id)
	if token != "" {
		request_url = request_url + "?token=" + token
	}

	result := &Msg_resp_app_version{}
	err := Get_(request_url, "", 30, result)
	if err != nil {
		return nil, errors.New("get app version error:" + err.Error() + " , package_id:" + strconv.Itoa(package_id))
	}
	if result.Meta_status <= 0 {
		return nil, errors.New("get app version error:" + result.Meta_message + " , package_id:" + strconv.Itoa(package_id))
	}

	//check min-allow version format
	if _, v_err := ParseVersion(result.Minimum_allow_version); v_err != nil {
		return nil, errors.New("minimum_allow_version format err, minimum_allow_version:" + result.Minimum_allow_version)
	}

	return result, nil
}
