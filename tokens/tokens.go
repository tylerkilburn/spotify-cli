package tokens

import (
	"log"
	"os"
	"spotify-cli/config"
	"time"

	"gopkg.in/yaml.v3"
)

const ACCESS_TOKEN_FILENAME = "spotify-cli-access.yaml"
const REFRESH_TOKEN_FILENAME = "spotify-cli-refresh.yaml"
const FILE_PERMISSIONS = 0666
const HOUR_IN_SECONDS = 3_600
const SIXTEEN_DAYS_IN_SECONDS = 1_382_400

type TokenWithExp struct {
	Token string `yaml:"token"`
	Exp   int64  `yaml:"exp"`
}

func GetAccessToken() string {
	return readToken(ACCESS_TOKEN_FILENAME)
}

func GetRefreshToken() string {
	return readToken(REFRESH_TOKEN_FILENAME)
}

func SetAccessToken(token string) {
	exp := getCurrentEpoch() + HOUR_IN_SECONDS
	writeToken(ACCESS_TOKEN_FILENAME, token, exp)
}

func SetRefreshToken(token string) {
	exp := getCurrentEpoch() + SIXTEEN_DAYS_IN_SECONDS
	writeToken(REFRESH_TOKEN_FILENAME, token, exp)
}

func getCurrentEpoch() int64 {
	return time.Now().Unix()
}

func readToken(filename string) string {
	var tokenWithExp TokenWithExp

	data, err := os.ReadFile(config.Get().Tokens.LocalDirectory + "/" + filename)
	if err != nil {
		log.Fatalf("Error reading data: %v", err)
		return ""
	}

	err = yaml.Unmarshal(data, &tokenWithExp)
	if err != nil {
		log.Fatalf("Error unmarshaling data: %v", err)
		return ""
	}

	if tokenWithExp.Exp < getCurrentEpoch() {
		return ""
	}

	return tokenWithExp.Token
}

func writeToken(filename string, token string, exp int64) {
	tokenWithExp := TokenWithExp{Token: token, Exp: exp}
	directory := config.Get().Tokens.LocalDirectory

	data, err := yaml.Marshal(tokenWithExp)
	if err != nil {
		log.Fatalf("Error marshaling data: %v", err)
	}

	err = os.MkdirAll(directory, FILE_PERMISSIONS)
	if err != nil {
		log.Fatalf("Error making directory: %v", err)
	}

	err = os.WriteFile(directory+"/"+filename, data, FILE_PERMISSIONS)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
}
