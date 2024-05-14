package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	uuid "github.com/google/uuid"
	"github.com/thomas-introini/pocket-cli/models"
	"github.com/thomas-introini/pocket-cli/utils"
)

const POCKET_URL = "https://getpocket.com"

var POCKET_CONSUMER_KEY = os.Getenv("POCKET_CONSUMER_KEY")

func GetRequestToken(consumerKey string, redirectURI string) (code string, state string, err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return
	}
	state = uuid.String()
	resp, err := http.PostForm(POCKET_URL+"/v3/oauth/request", url.Values{"consumer_key": {consumerKey}, "state": {state}, "redirect_uri": {redirectURI}})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	url, err := url.Parse("/?" + string(body))
	if err != nil {
		return
	}
	values := url.Query()
	if values.Get("state") != uuid.String() {
		err = errors.New("state does not match")
		return
	}
	code = values.Get("code")
	if code == "" {
		err = errors.New("code is empty")
		return
	}
	return
}

func GetAccesToken(consumerKey string, state string, code string) (accessToken string, username string, err error) {
	resp, err := http.PostForm(POCKET_URL+"/v3/oauth/authorize", url.Values{"consumer_key": {consumerKey}, "code": {code}})
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("get access token: status code " + resp.Status)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	url, err := url.Parse("/?" + string(body))
	if err != nil {
		return
	}
	values := url.Query()
	if values.Get("state") != state {
		err = errors.New("get access token: state does not match")
		return
	}
	accessToken = values.Get("access_token")
	if accessToken == "" {
		err = errors.New("code is empty")
		return
	}
	username = values.Get("username")
	if accessToken == "" {
		err = errors.New("username is empty")
		return
	}
	return

}

func OpenAuthorizationURL(requestToken string, redirectURI string) error {
	err := utils.OpenInBrowser(POCKET_URL + "/auth/authorize?request_token=" + requestToken + "&redirect_uri=" + redirectURI)
	if err != nil {
		return err
	}
	return nil
}

type PocketSavesResponse struct {
	Since float64
	Saves []models.PocketSave
}

func GetAllPocketSaves(accessToken string) (PocketSavesResponse, error) {
	body := map[string]any{
		"consumer_key": POCKET_CONSUMER_KEY,
		"access_token": accessToken,
		"state":        "all",
		"sort":         "newest",
		"detailType":   "simple",
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return PocketSavesResponse{}, err
	}

	response, err := http.Post(POCKET_URL+"/v3/get", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return PocketSavesResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return PocketSavesResponse{}, errors.New("could not retrieve saves: " + response.Status)
	}

	var jsonResponse map[string]interface{}

	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		return PocketSavesResponse{}, err
	}
	since := jsonResponse["since"].(float64)
	list := jsonResponse["list"].(map[string]interface{})
	saves := make([]models.PocketSave, 0)
	for _, save := range list {
		save := save.(map[string]interface{})
		updatedOn, err := strconv.Atoi(save["time_updated"].(string))
		if err != nil {
			return PocketSavesResponse{}, err
		}
		e := save["excerpt"]
		excerpt := ""
		if e != nil {
			excerpt = e.(string)
		}
		saves = append(saves, models.PocketSave{
			Id:              save["item_id"].(string),
			SaveTitle:       save["given_title"].(string),
			Url:             save["given_url"].(string),
			SaveDescription: excerpt,
			UpdatedOn:       int32(updatedOn),
		})
	}

	return PocketSavesResponse{
		Since: since,
		Saves: saves,
	}, nil
}
