package kopia

import (
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd/util"
)

var (
	namespace          string
	secretFilePath     string
	backupLocationName string
	backupLocationFile string
	volumeBackupName   string
	kopiaRepo          string
)

// NewCommand returns a kopia command wrapper
func NewCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "kopia_executor",
		Short: "a command executor for long running kopia commands",
	}

	cmds.PersistentFlags().StringVar(&backupLocationName, "backup-location", "", "Name of the BackupLocation object, used for authentication")
	cmds.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace for this command")
	cmds.PersistentFlags().StringVar(&backupLocationFile, "backup-location-file", "", "Path to the BackupLocation object, used for authentication")
	cmds.PersistentFlags().StringVar(&kopiaRepo, "repository", "", "Name of the kopia repository. If provided it will overwrite the BackupLocation one")
	cmds.PersistentFlags().StringVarP(&secretFilePath, "secret-file-path", "s", "", "Path of the secret file used for locking/unlocking kopia reposiories")

	cmds.AddCommand(
		newBackupCommand(),
		//newRestoreCommand(),
	)
	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		util.CheckErr(err)
		return nil
	}

	return cmds
}
