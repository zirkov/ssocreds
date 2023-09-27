package ssocreds

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func getCallerIdentity(cfg aws.Config) (*sts.GetCallerIdentityOutput, error) {
	client := sts.NewFromConfig(cfg)
	input := &sts.GetCallerIdentityInput{}
	return client.GetCallerIdentity(context.Background(), input)
}

func AccountAndPermissionSet(cfg aws.Config) (string, string, error) {
	const (
		roleSeparator = "_"
		arnSeparator  = "/"
	)

	output, err := getCallerIdentity(cfg)

	if err != nil {
		return "", "", err
	}

	account := *output.Account
	arn := *output.Arn

	arnParts := strings.Split(arn, arnSeparator)

	if len(arnParts) != 3 {
		return "", "", fmt.Errorf("cannot parse arn: %s", arn)
	}

	role := arnParts[1]

	roleParts := strings.Split(role, roleSeparator)

	if len(roleParts) < 3 {
		return "", "", fmt.Errorf("cannot parse role: %s", role)
	}

	roleName := strings.Join(roleParts[1:len(roleParts)-1], roleSeparator)
	return account, roleName, nil
}
