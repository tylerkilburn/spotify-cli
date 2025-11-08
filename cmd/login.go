package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"spotify-cli/utils"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const PORT = 4202

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a C LI library for Go that empowers applications.
This ap:Wplication is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serving login")

		http.HandleFunc("/login", httpLogin)
		http.HandleFunc("/callback", httpCallback)

		http.ListenAndServe(":"+fmt.Sprint(PORT), nil)
		fmt.Println("Server starting on port " + fmt.Sprint(PORT) + "...")

		time.Sleep(1 * time.Second)
	},
}

const CLIENT_ID = ""
const CLIENT_SECRET = ""

var scope = []string{"user-read-private", "user-read-email"}
var redirect_uri = "http://localhost:" + fmt.Sprint(PORT) + "/callback"

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func httpLogin(w http.ResponseWriter, req *http.Request) {
	baseURL := "https://accounts.spotify.com/authorize" // A simple service to test GET requests

	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	q := u.Query()
	q.Add("response_type", "code")
	q.Add("client_id", CLIENT_ID)
	q.Add("scope", strings.Join(scope, " "))
	q.Add("redirect_uri", redirect_uri)
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
	formData.Set("redirect_uri", redirect_uri)
	formData.Set("grant_type", "authorization_code")

	// Create a new HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the Content-Type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add custom headers
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(CLIENT_ID + ":" + CLIENT_SECRET))
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

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
}
