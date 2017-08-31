package ebs

import (
	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
)

const (
	// ConfigEBS is a config key.
	ConfigEBS = "ebs"

	// ConfigEBSAccessKey is a config key.
	ConfigEBSAccessKey = ConfigEBS + "." + AccessKey

	// ConfigEBSSecretKey is a config key.
	ConfigEBSSecretKey = ConfigEBS + "." + SecretKey

	// ConfigEBSRegion is a config key.
	ConfigEBSRegion = ConfigEBS + "." + Region

	// ConfigEBSEndpoint is a config key.
	ConfigEBSEndpoint = ConfigEBS + "." + Endpoint

	// ConfigEBSMaxRetries is a config key.
	ConfigEBSMaxRetries = ConfigEBS + "." + MaxRetries

	// ConfigEBSTag is a config key.
	ConfigEBSTag = ConfigEBS + "." + Tag

	// ConfigEBSRexrayTag is a config key.
	ConfigEBSRexrayTag = ConfigEBS + ".rexrayTag"

	// ConfigEBSKmsKeyID is a config key.
	ConfigEBSKmsKeyID = ConfigEBS + "." + KmsKeyID

	// ConfigEC2 is a config key.
	ConfigEC2 = "ec2"

	// ConfigEC2AccessKey is a config key.
	ConfigEC2AccessKey = ConfigEC2 + "." + AccessKey

	// ConfigEC2SecretKey is a config key.
	ConfigEC2SecretKey = ConfigEC2 + "." + SecretKey

	// ConfigEC2Region is a config key.
	ConfigEC2Region = ConfigEC2 + "." + Region

	// ConfigEC2Endpoint is a config key.
	ConfigEC2Endpoint = ConfigEC2 + "." + Endpoint

	// ConfigEC2MaxRetries is a config key.
	ConfigEC2MaxRetries = ConfigEC2 + "." + MaxRetries

	// ConfigEC2Tag is a config key.
	ConfigEC2Tag = ConfigEC2 + "." + Tag

	// ConfigEC2RexrayTag is a config key.
	ConfigEC2RexrayTag = ConfigEC2 + ".rexrayTag"

	// ConfigEC2KmsKeyID is a config key.
	ConfigEC2KmsKeyID = ConfigEC2 + "." + KmsKeyID

	// ConfigAWS is a config key.
	ConfigAWS = "aws"

	// ConfigAWSAccessKey is a config key.
	ConfigAWSAccessKey = ConfigAWS + "." + AccessKey

	// ConfigAWSSecretKey is a config key.
	ConfigAWSSecretKey = ConfigAWS + "." + SecretKey

	// ConfigAWSRegion is a config key.
	ConfigAWSRegion = ConfigAWS + "." + Region

	// ConfigAWSEndpoint is a config key.
	ConfigAWSEndpoint = ConfigAWS + "." + Endpoint

	// ConfigAWSMaxRetries is a config key.
	ConfigAWSMaxRetries = ConfigAWS + "." + MaxRetries

	// ConfigAWSTag is a config key.
	ConfigAWSTag = ConfigAWS + "." + Tag

	// ConfigAWSRexrayTag is a config key.
	ConfigAWSRexrayTag = ConfigAWS + ".rexrayTag"

	// ConfigAWSKmsKeyID is a config key.
	ConfigAWSKmsKeyID = ConfigAWS + "." + KmsKeyID
)

// BackCompat ensures keys can be used from old configurations.
func BackCompat(config gofig.Config) {
	ec2Checks := [][]string{
		{ConfigEBSAccessKey, ConfigEC2AccessKey},
		{ConfigEBSSecretKey, ConfigEC2SecretKey},
		{ConfigEBSRegion, ConfigEC2Region},
		{ConfigEBSEndpoint, ConfigEC2Endpoint},
		{ConfigEBSMaxRetries, ConfigEC2MaxRetries},
		{ConfigEBSTag, ConfigEC2Tag},
		{ConfigEBSRexrayTag, ConfigEC2RexrayTag},
		{ConfigEBSKmsKeyID, ConfigEC2KmsKeyID},
	}
	for _, check := range ec2Checks {
		if !config.IsSet(check[0]) && config.IsSet(check[1]) {
			log.Debug(config.Get(check[1]))
			config.Set(check[0], config.Get(check[1]))
		}
	}

	awsChecks := [][]string{
		{ConfigEBSAccessKey, ConfigAWSAccessKey},
		{ConfigEBSSecretKey, ConfigAWSSecretKey},
		{ConfigEBSRegion, ConfigAWSRegion},
		{ConfigEBSEndpoint, ConfigAWSEndpoint},
		{ConfigEBSMaxRetries, ConfigAWSMaxRetries},
		{ConfigEBSTag, ConfigAWSTag},
		{ConfigEBSRexrayTag, ConfigAWSRexrayTag},
		{ConfigEBSKmsKeyID, ConfigAWSKmsKeyID},
	}
	for _, check := range awsChecks {
		if !config.IsSet(check[0]) && config.IsSet(check[1]) {
			log.Debug(config.Get(check[1]))
			config.Set(check[0], config.Get(check[1]))
		}
	}
}
