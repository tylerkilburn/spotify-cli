package spot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"spotify-cli/config"
	"spotify-cli/tokens"
	"spotify-cli/utils"
	"strings"
	"time"
)

const SPOTIFY_BASE_URL = "https://accounts.spotify.com/authorize" // A simple service to test GET requests

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
	u, err := url.Parse(SPOTIFY_BASE_URL)
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
	//   var code = req.query.code || null;
	//   var state = req.query.state || null;
	queryParams := req.URL.Query()

	// Access a specific parameter by key
	code := queryParams.Get("code")
	// state := queryParams.Get("state")

	// http.Post("https://accounts.spotify.com/api/token",http)

	// Set form data
	formData := url.Values{}
	formData.Set("code", code)
	formData.Set("redirect_uri", config.Get().Http.RedirectUri)
	formData.Set("grant_type", "authorization_code")

	// Create a new HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the Content-Type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add custom headers
	clientCreds := config.Get().Spotify.ClientId + ":" + config.Get().Spotify.ClientSecret
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(clientCreds))
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	// Create an HTTP client (you can customize it with timeouts, etc.)
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close() // Close the response body when the function exits

	// Read and print the response body
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

	fmt.Printf("AccesToken: %s\n", tokenResponse.AccessToken)
	fmt.Printf("RefreshToken: %s\n", tokenResponse.RefreshToken)
}
