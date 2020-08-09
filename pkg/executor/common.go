package executor

import (
	"fmt"
	"io/ioutil"
	"os"

	storkapi "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	storkops "github.com/portworx/sched-ops/k8s/stork"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	amazonS3Endpoint      = "s3.amazonaws.com"
	googleAccountFilePath = "/root/.gce_credentials"
)

// Repository contains information used to connect the repository.
type Repository struct {
	// Name is a repository name without an url address.
	Name string
	// Path is a full repository name.
	Path string
	// AuthEnv is a set of environment variables used for authentication.
	AuthEnv []string
}

// ParseBackupLocation parses the provided backup location and returns the repository name
func ParseBackupLocation(repoName, name, namespace, filePath string) (*Repository, error) {
	backupLocation, err := readBackupLocation(name, namespace, filePath)
	if err != nil {
		return nil, err
	}

	switch backupLocation.Location.Type {
	case storkapi.BackupLocationS3:
		return parseS3(repoName, backupLocation.Location)
	case storkapi.BackupLocationAzure:
		return parseAzure(repoName, backupLocation.Location)
	case storkapi.BackupLocationGoogle:
		return parseGce(repoName, backupLocation.Location)
	}
	return nil, fmt.Errorf("unsupported backup location: %v", backupLocation.Location.Type)
}

func readBackupLocation(name, namespace, filePath string) (*storkapi.BackupLocation, error) {
	if name != "" {
		if namespace == "" {
			namespace = "default"
		}
		return storkops.Instance().GetBackupLocation(name, namespace)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	out := &storkapi.BackupLocation{}
	if err = yaml.NewYAMLOrJSONDecoder(f, 1024).Decode(out); err != nil {
		return nil, err
	}

	return out, nil
}

func parseS3(repoName string, backupLocation storkapi.BackupLocationItem) (*Repository, error) {
	if backupLocation.S3Config == nil {
		return nil, fmt.Errorf("failed to parse s3 config from BackupLocation")
	}

	envs := make([]string, 0)
	envs = append(envs, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", backupLocation.S3Config.AccessKeyID))
	envs = append(envs, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", backupLocation.S3Config.SecretAccessKey))
	if backupLocation.S3Config.Region != "" {
		envs = append(envs, fmt.Sprintf("AWS_REGION=%s", backupLocation.S3Config.Region))
	}

	if repoName == "" {
		repoName = backupLocation.Path
	}
	return &Repository{
		Name:    repoName,
		Path:    fmt.Sprintf("s3:%s/%s", backupLocation.S3Config.Endpoint, repoName),
		AuthEnv: envs,
	}, nil
}

func parseAzure(repoName string, backupLocation storkapi.BackupLocationItem) (*Repository, error) {
	if backupLocation.AzureConfig == nil {
		return nil, fmt.Errorf("failed to parse azure config from BackupLocation")
	}
	envs := make([]string, 0)
	envs = append(envs, fmt.Sprintf("AZURE_ACCOUNT_NAME=%s", backupLocation.AzureConfig.StorageAccountName))
	envs = append(envs, fmt.Sprintf("AZURE_ACCOUNT_KEY=%s", backupLocation.AzureConfig.StorageAccountKey))

	if repoName == "" {
		repoName = backupLocation.Path
	}
	return &Repository{
		Name:    repoName,
		Path:    "azure:" + repoName + "/",
		AuthEnv: envs,
	}, nil
}

func parseGce(repoName string, backupLocation storkapi.BackupLocationItem) (*Repository, error) {
	if backupLocation.GoogleConfig == nil {
		return nil, fmt.Errorf("failed to parse google config from BackupLocation")
	}

	if err := ioutil.WriteFile(
		googleAccountFilePath,
		[]byte(backupLocation.GoogleConfig.AccountKey),
		0644,
	); err != nil {
		return nil, fmt.Errorf("failed to parse google account key: %v", err)
	}

	envs := make([]string, 0)
	envs = append(envs, fmt.Sprintf("GOOGLE_PROJECT_ID=%s", backupLocation.GoogleConfig.ProjectID))
	envs = append(envs, fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", googleAccountFilePath))

	if repoName == "" {
		repoName = backupLocation.Path
	}
	return &Repository{
		Name:    repoName,
		Path:    "gs:" + repoName + "/",
		AuthEnv: envs,
	}, nil
}
