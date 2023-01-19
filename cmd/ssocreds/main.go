package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/winebarrel/ssocreds"
	"github.com/winebarrel/ssocreds/utils"
)

func init() {
	log.SetFlags(0)
}

var allowedFormats = []string{
	"env",
	"json",
}

const (
	defaultFormat = "env"
)

func main() {
	var formatPtr = flag.String("format", defaultFormat, fmt.Sprintf("Output format, one of (%v)", allowedFormats))
	var profilePtr = flag.String("profile", "", "Profile to use, same value as passed to AWS CLI --profile")

	flag.Parse()

	format := *formatPtr
	if !utils.Contains(allowedFormats, format) {
		log.Fatalf("invalid format: %s (allowed formats: %v)", format, allowedFormats)
	}

	var profile string
	if *profilePtr != "" {
		profile = *profilePtr
	} else {
		profile = os.Getenv("AWS_PROFILE")
	}

	if profile == "" {
		log.Fatal("AWS_PROFILE is not set and no profile passed as --profile")
	}

	startUrl, err := ssocreds.SsoStartUrl(profile)

	if err != nil {
		log.Fatal(err)
	}

	accessToken, region, err := ssocreds.LastAccessTokenAndRegion(startUrl)

	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	account, permissionSet, err := ssocreds.AccountAndPermissionSet(cfg)

	if err != nil {
		log.Fatal(err)
	}

	accessKeyId, secretAccessKey, sessionToken, err := ssocreds.SsoCredentials(cfg, account, permissionSet, accessToken, region)

	if err != nil {
		log.Fatal(err)
	}

	switch format {
	case "env":
		printEnv(accessKeyId, secretAccessKey, sessionToken)
	case "json":
		printJson(accessKeyId, secretAccessKey, sessionToken)
	default:
		log.Panicf("invalid format: %s,", format)
	}

}

func printEnv(accessKeyId, secretAccessKey, sessionToken string) {
	fmt.Printf("export AWS_ACCESS_KEY_ID='%s'\n", accessKeyId)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY='%s'\n", secretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN='%s'\n", sessionToken)
}

func printJson(accessKeyId, secretAccessKey, sessionToken string) {
	creds := map[string]string{
		"accessKeyId":     accessKeyId,
		"secretAccessKey": secretAccessKey,
		"sessionToken":    sessionToken,
	}

	output, _ := json.MarshalIndent(creds, "", "  ")
	fmt.Println(string(output))
}
