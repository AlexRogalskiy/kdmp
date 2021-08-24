package kopia

import (
	"fmt"
	"time"

	storkv1 "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/libopenstorage/stork/pkg/objectstore"
	"github.com/portworx/kdmp/pkg/executor"
	"github.com/portworx/kdmp/pkg/kopia"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gocloud.dev/blob"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/cmd/util"
)

const (
	progressCheckInterval = 5 * time.Second
	genericBackupDir      = "generic-backup"
)

func newBackupCommand() *cobra.Command {
	var (
		sourcePath     string
		sourcePathGlob string
	)
	backupCommand := &cobra.Command{
		Use:   "backup",
		Short: "Start a kopia backup",
		Run: func(c *cobra.Command, args []string) {
			srcPath, err := executor.GetSourcePath(sourcePath, sourcePathGlob)
			if err != nil {
				util.CheckErr(err)
				return
			}

			executor.HandleErr(runBackup(srcPath))
		},
	}
	backupCommand.Flags().StringVar(&sourcePath, "source-path", "", "Source for kopia backup")
	backupCommand.Flags().StringVar(&sourcePathGlob, "source-path-glob", "", "The regexp should match only one path that will be used for backup")
	backupCommand.Flags().StringVar(&volumeBackupName, "volume-backup-name", "", "Provided VolumeBackup CRD will be updated with the latest backup progress details")
	return backupCommand
}

func runBackup(sourcePath string) error {
	// Parse using the mounted secrets
	fn := "runBackup"
	repo, err := executor.ParseCloudCred()
	repo.Name = frameBackupPath()

	if err != nil {
		if statusErr := executor.WriteVolumeBackupStatus(
			&executor.Status{LastKnownError: err},
			volumeBackupName,
			namespace,
		); statusErr != nil {
			return statusErr
		}
		return fmt.Errorf("parse backuplocation: %s", err)
	}
	if volumeBackupName != "" {
		if err = executor.CreateVolumeBackup(
			volumeBackupName,
			namespace,
			repo.Name,
			credentials,
		); err != nil {
			logrus.Errorf("%s: %v", fn, err)
			return err
		}
	}
	// kopia doesn't have a way to know if repository is already initialized.
	// Repository create needs to run only first time.
	// Check if kopia.repository exists
	exists, err := isRepositoryExists(repo)
	if err != nil {
		errMsg := fmt.Sprintf("repository exists check failed: %v", err)
		logrus.Errorf("%s: %v", fn, errMsg)
		return fmt.Errorf("%s: %v", errMsg, err)
	}

	if !exists {
		if err = runKopiaCreateRepo(repo); err != nil {
			errMsg := "repo creation failed"
			logrus.Errorf("%s: %v", fn, errMsg)
			return fmt.Errorf("%s: %v", errMsg, err)
		}
	}

	if err = runKopiaRepositoryConnect(repo); err != nil {
		errMsg := fmt.Sprintf("repository connect failed: %v", err)
		logrus.Errorf("%s: %v", fn, errMsg)
		return fmt.Errorf(errMsg)
	}

	if err = runKopiaBackup(repo, sourcePath); err != nil {
		errMsg := fmt.Sprintf("backup failed: %v", err)
		logrus.Errorf("%s: %v", fn, errMsg)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func buildStorkBackupLocation(repository *executor.Repository) *storkapi.BackupLocation {
	//var backupType storkapi.BackupLocationType
	backupLocation := &storkapi.BackupLocation{
		ObjectMeta: meta.ObjectMeta{},
		Location:   storkapi.BackupLocationItem{},
	}
	switch repository.Type {
	case storkapi.BackupLocationS3:
		backupLocation.Location.S3Config = &storkapi.S3Config{
			AccessKeyID:      repository.S3Config.AccessKeyID,
			SecretAccessKey:  repository.S3Config.SecretAccessKey,
			Endpoint:         repository.S3Config.Endpoint,
			Region:           repository.S3Config.Endpoint,
		}
	}
	backupLocation.Location.Path = repository.Path
	backupLocation.ObjectMeta.Name = repository.Name
	backupLocation.Location.Type = storkapi.BackupLocationS3

	return backupLocation
}

func populateS3AccessDetails(initCmd *kopia.Command, repository *executor.Repository) *kopia.Command {
	// kopia is not honouring env variabels set in the pod so passing them as flags
	initCmd.AddArg("--endpoint")
	initCmd.AddArg(repository.S3Config.Endpoint)
	initCmd.AddArg("--access-key")
	initCmd.AddArg(repository.S3Config.AccessKeyID)
	initCmd.AddArg("--secret-access-key")
	initCmd.AddArg(repository.S3Config.SecretAccessKey)

	return initCmd
}

func runKopiaCreateRepo(repository *executor.Repository) error {
	logrus.Infof("kopia repository creation started...")
	initCmd, err := kopia.GetCreateCommand(repository.Path, repository.Name, repository.Password, string(repository.Type))
	if err != nil {
		return err
	}
	// TODO: Add for other storage providers
	initCmd = populateS3AccessDetails(initCmd, repository)
	initExecutor := kopia.NewCreateExecutor(initCmd)
	if err := initExecutor.Run(); err != nil {
		err = fmt.Errorf("failed to run kopia init command: %v", err)
		return err
	}
	for {
		time.Sleep(progressCheckInterval)
		status, err := initExecutor.Status()
		if err != nil {
			return err
		}
		if status.LastKnownError != nil {
			if status.LastKnownError != kopia.ErrAlreadyRepoExist {
				return status.LastKnownError
			}
			status.LastKnownError = nil
		}

		if err = executor.WriteVolumeBackupStatus(
			status,
			volumeBackupName,
			namespace,
		); err != nil {
			logrus.Errorf("failed to write a VolumeBackup status: %v", err)
			continue
		}
		if status.Done {
			break
		}
	}

	logrus.Infof("kopia repository creation successful...")
	return nil
}

func runKopiaBackup(repository *executor.Repository, sourcePath string) error {
	logrus.Infof("kopia backup started...")
	backupCmd, err := kopia.GetBackupCommand(
		repository.Path,
		repository.Name,
		repository.Password,
		string(repository.Type),
		sourcePath,
	)
	if err != nil {
		return err
	}
	// This is needed to handle case where aftert kopia repo create it was successful
	// the pod got terminated. Now user trigerrs another backup, so we need to pass
	// credentials for "snapshot create".
	//backupCmd = populateS3AccessDetails(backupCmd, repository)
	initExecutor := kopia.NewBackupExecutor(backupCmd)
	if err := initExecutor.Run(); err != nil {
		err = fmt.Errorf("failed to run backup command: %v", err)
		return err
	}

	for {
		time.Sleep(progressCheckInterval)
		status, err := initExecutor.Status()
		if err != nil {
			return err
		}
		if status.LastKnownError != nil {
			return status.LastKnownError
		}

		if err = executor.WriteVolumeBackupStatus(
			status,
			volumeBackupName,
			namespace,
		); err != nil {
			logrus.Errorf("failed to write a VolumeBackup status: %v", err)
			continue
		}
		if status.Done {
			break
		}
	}

	logrus.Infof("kopia backup successful...")

	return nil
}

func runKopiaRepositoryConnect(repository *executor.Repository) error {
	connectCmd, err := kopia.GetConnectCommand(repository.Path, repository.Name, repository.Password, string(repository.Type))
	if err != nil {
		return err
	}
	// TODO: Add for other storage providers
	connectCmd = populateS3AccessDetails(connectCmd, repository)
	initExecutor := kopia.NewConnectExecutor(connectCmd)
	if err := initExecutor.Run(); err != nil {
		err = fmt.Errorf("failed to run repository connect  command: %v", err)
		return err
	}
	//TODO: Temp commented out
	for {
		time.Sleep(progressCheckInterval)
		status, err := initExecutor.Status()
		if err != nil {
			return err
		}
		if status.LastKnownError != nil {
			return status.LastKnownError
		}
		if status.Done {
			break
		}
	}

	return nil
}

// Under backuplocaiton path, following path would be created
// <bucket>/generic-backup/<ns - pvc>
func frameBackupPath() string {
	return genericBackupDir + "/" + kopiaRepo + "/"
}

func buildStorkBackupLocation(repository *executor.Repository) (*storkv1.BackupLocation, error) {
	var backupType storkv1.BackupLocationType
	backupLocation := &storkv1.BackupLocation{
		ObjectMeta: metav1.ObjectMeta{},
		Location:   storkv1.BackupLocationItem{},
	}

	switch repository.Type {
	case storkv1.BackupLocationS3:
		backupType = storkv1.BackupLocationS3
		backupLocation.Location.S3Config = &storkv1.S3Config{
			AccessKeyID:     repository.S3Config.AccessKeyID,
			SecretAccessKey: repository.S3Config.SecretAccessKey,
			Endpoint:        repository.S3Config.Endpoint,
			Region:          repository.S3Config.Region,
		}
	}

	backupLocation.Location.Path = repository.Path
	backupLocation.ObjectMeta.Name = repository.Name
	backupLocation.Location.Type = backupType
	return backupLocation, nil
}

func isRepositoryExists(repository *executor.Repository) (bool, error) {
	bl, err := buildStorkBackupLocation(repository)
	if err != nil {
		logrus.Errorf("%v", err)
		return false, err
	}
	bucket, err := objectstore.GetBucket(bl)
	if err != nil {
		logrus.Errorf("err: %v", err)
		return false, err
	}
	bucket = blob.PrefixedBucket(bucket, repository.Name)
	exists, err := bucket.Exists(context.TODO(), "kopia.repository")
	if err != nil {
		logrus.Errorf("%v", err)
		return false, err
	}
	if exists {
		logrus.Infof("kopia.repository exists")
	} else {
		logrus.Infof("kopia.repository doesn't exists")
	}
	return exists, nil
}