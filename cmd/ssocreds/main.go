package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/winebarrel/ssocreds"
	"github.com/winebarrel/ssocreds/utils"
)

func init() {
	log.SetFlags(0)
}

type SsoAccessData struct {
	awsProfile, accessKeyId, secretAccessKey, sessionToken string
}

var allowedFormats = []string{
	"env",
	"json",
}

const (
	defaultFormat = "env"
	awsCredPath   = ".aws/credentials"
)

func main() {
	var formatPtr = flag.String("format", defaultFormat, fmt.Sprintf("Output format, one of (%v)", allowedFormats))
	var profilePtr = flag.String("profile", "", "Profile to use, same value as passed to AWS CLI --profile")
	var forcePtr = flag.Bool("force", false, "Force cleanup of old credentials and setting of AWS_PROFILE")

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

	force := *forcePtr
	if force {
		awsCredFile := filepath.Join(utils.HomeDir(), awsCredPath)
		err := os.Rename(awsCredFile, filepath.Join(utils.HomeDir(), ".aws", "old.credentials.backup"))
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	startUrl, err := ssocreds.SsoStartUrl(profile)

	if err != nil {
		log.Fatal(err)
	}

	accessToken, region, err := ssocreds.LastAccessTokenAndRegion(startUrl)

	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigProfile(profile))

	if err != nil {
		log.Fatal(err)
	}

	account, permissionSet, err := ssocreds.AccountAndPermissionSet(cfg)

	if err != nil {
		log.Fatal(err)
	}

	var accessData = SsoAccessData{}
	if force {
		accessData.awsProfile = profile
	}

	accessData.accessKeyId, accessData.secretAccessKey, accessData.sessionToken, err =
		ssocreds.SsoCredentials(cfg, account, permissionSet, accessToken, region)

	if err != nil {
		log.Fatal(err)
	}

	switch format {
	case "env":
		printEnv(accessData, force)
	case "json":
		printJson(accessData)
	default:
		log.Panicf("invalid format: %s,", format)
	}

}

func printEnv(accessData SsoAccessData, force bool) {
	if force {
		fmt.Printf("export AWS_PROFILE='%s'\n", accessData.awsProfile)
	}
	fmt.Printf("export AWS_ACCESS_KEY_ID='%s'\n", accessData.accessKeyId)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY='%s'\n", accessData.secretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN='%s'\n", accessData.sessionToken)
}

func printJson(accessData SsoAccessData) {
	creds := map[string]string{
		"accessKeyId":     accessData.accessKeyId,
		"secretAccessKey": accessData.secretAccessKey,
		"sessionToken":    accessData.sessionToken,
	}

	output, _ := json.MarshalIndent(creds, "", "  ")
	fmt.Println(string(output))
}
