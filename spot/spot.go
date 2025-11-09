package spot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"spotify-cli/config"
	"spotify-cli/tokens"
	"spotify-cli/utils"
	"strings"
	"time"
)

const AUTHORIZATION_CODE = "authorization_code"
const REFRESH_TOKEN = "refresh_token"
const SPOTIFY_AUTHORIZE_URL = "https://accounts.spotify.com/authorize" // A simple service to test GET requests
const SPOTIFY_PLAY_URL = "https://api.spotify.com/v1/me/player/play"

func PlayPlaylist(playlistId string) {
	payload, err := json.Marshal(map[string]string{
		"context_uri": "spotify:playlist:" + playlistId,
	})
	if err != nil {
		log.Fatalf("Error Marshaling play body: %v", err)
	}

	accessToken := tokens.GetAccessToken()
	if accessToken == "" {
		accessToken = refreshToken()
	}

	req, err := http.NewRequest(http.MethodPut, SPOTIFY_PLAY_URL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error when calling Spotify play endpoint: %v", err)
	}
	defer response.Body.Close()

	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error invoking Spotify play: %v", err)
	}
}

func ServeLocalLoginPageAndAuthUser() {
	httpConfig := config.Get().Http

	fmt.Println("serving login")
	fmt.Println(httpConfig.LoginEndpoint)
	fmt.Println(httpConfig.RedirectEndpoint)
	fmt.Println(httpConfig.RedirectUri)

	http.HandleFunc(httpConfig.LoginEndpoint, httpLogin)
	http.HandleFunc(httpConfig.RedirectEndpoint, httpCallback)

	http.ListenAndServe(":"+fmt.Sprint(httpConfig.Port), nil)
	fmt.Println("Server starting on port " + fmt.Sprint(httpConfig.Port) + "...")

	time.Sleep(1 * time.Second)
}

func httpLogin(w http.ResponseWriter, req *http.Request) {
	u, err := url.Parse(SPOTIFY_AUTHORIZE_URL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	q := u.Query()
	q.Add("response_type", "code")
	q.Add("client_id", config.Get().Spotify.ClientId)
	q.Add("scope", strings.Join(config.Get().Spotify.Scope, " "))
	q.Add("redirect_uri", config.Get().Http.RedirectUri)
	q.Add("state", utils.GenerateRandomString(16))

	u.RawQuery = q.Encode()
	fmt.Println("Full URL with query parameters:", u.String())

	http.Redirect(w, req, u.String(), http.StatusTemporaryRedirect)
}

func httpCallback(w http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	code := queryParams.Get("code")
	exchangeAuthoriztionCode(code)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged into Spotify, you may close this browser tab!"))

	asyncExit()
}

func asyncExit() {
	ch := make(chan int, 0)
	go os.Exit(0) // Find better way for graceful http shutdown
	defer close(ch)
}

type TokenApiParams struct {
	grantType    string
	code         string
	redirectUri  string
	refreshToken string
}

func exchangeAuthoriztionCode(code string) {
	exchangeToken(&TokenApiParams{grantType: AUTHORIZATION_CODE, code: code, redirectUri: config.Get().Http.RedirectUri})
}

func refreshToken() string {
	return exchangeToken(&TokenApiParams{grantType: REFRESH_TOKEN, refreshToken: tokens.GetRefreshToken()})
}

func exchangeToken(tokenApiParams *TokenApiParams) string {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	if tokenApiParams.grantType == "authorization_code" {
		formData.Set("code", tokenApiParams.code)
		formData.Set("redirect_uri", tokenApiParams.redirectUri)
	} else {
		formData.Set("refresh_token", tokenApiParams.refreshToken)
	}

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	clientCreds := config.Get().Spotify.ClientId + ":" + config.Get().Spotify.ClientSecret
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(clientCreds))
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close() // Close the response body when the function exits

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	type TokenResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		log.Fatalf("Error unmarshaling response body: %v", err)
	}

	tokens.SetAccessToken(tokenResponse.AccessToken)
	tokens.SetRefreshToken(tokenResponse.RefreshToken)

	return tokenResponse.AccessToken
}
