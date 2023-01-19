package ssocreds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/winebarrel/ssocreds/utils"
)

const (
	awsSsoCacheDir = ".aws/sso/cache"
)

func awsCache() ([]map[string]string, error) {
	pattern := filepath.Join(utils.HomeDir(), awsSsoCacheDir, "*.json")
	files, err := filepath.Glob(pattern)

	if err != nil {
		return nil, err
	}

	caches := []map[string]string{}

	for _, f := range files {
		raw, err := ioutil.ReadFile(f)

		if err != nil {
			return nil, err
		}

		j := map[string]string{}
		err = json.Unmarshal(raw, &j)

		if err != nil {
			panic(err)
		}

		caches = append(caches, j)
	}

	return caches, nil
}

type awsSsoCahce struct {
	startUrl    string
	accessToken string
	region      string
	expiresAt   time.Time
}

func awsSsoCaches(caches []map[string]string) []*awsSsoCahce {
	ssoCaches := []*awsSsoCahce{}

	for _, c := range caches {
		startUrl, ok := c["startUrl"]

		if !ok {
			continue
		}

		accessToken, ok := c["accessToken"]

		if !ok {
			continue
		}

		region, ok := c["region"]

		if !ok {
			continue
		}

		expiresAt, ok := c["expiresAt"]

		if !ok {
			continue
		}

		t, err := time.Parse(time.RFC3339, expiresAt)

		if err != nil {
			panic(err)
		}

		ssoCaches = append(ssoCaches, &awsSsoCahce{
			startUrl:    startUrl,
			accessToken: accessToken,
			region:      region,
			expiresAt:   t,
		})
	}

	return ssoCaches
}

func LastAccessTokenAndRegion(startUrl string) (string, string, error) {
	caches, err := awsCache()
	ssoCaches := awsSsoCaches(caches)

	if err != nil {
		return "", "", err
	}

	var accessToken string
	var region string
	expiresAt := time.Time{}

	for _, c := range ssoCaches {
		if c.expiresAt.Before(time.Now()) {
			continue
		}

		if c.startUrl == startUrl && c.expiresAt.After(expiresAt) {
			accessToken = c.accessToken
			region = c.region
		}
	}

	if accessToken == "" {
		return "", "", fmt.Errorf("access token not found, try `aws sso login`")
	}

	return accessToken, region, nil
}
