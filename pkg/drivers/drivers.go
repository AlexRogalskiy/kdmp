package drivers

import "fmt"

// Known drivers.
const (
	Rsync         = "rsync"
	ResticBackup  = "resticbackup"
	ResticRestore = "resticrestore"
)

// Driver labels.
const (
	DriverNameLabel = "kdmp.portworx.com/driver-name"
)

const (
	// OpenshiftSCCAnnotation used to set a openshift securit context contraint.
	OpenshiftSCCAnnotation = "openshift.io/scc"
)

const (
	// TransferProgressCompleted is a status for a data transfer.
	TransferProgressCompleted float64 = 100
)

// Common parameters for restic secret.
const (
	SecretKey   = "secret"
	SecretValue = "resticsecret"
	SecretMount = "/tmp/resticsecret"
)

// Driver job options.
const (
	RsyncFlags                   = "KDMP_RSYNC_FLAGS"
	RsyncOpenshiftSCC            = "KDMP_RSYNC_OPENSHIFT_SCC"
	RsyncImageKey                = "KDMP_RSYNC_IMAGE"
	RsyncImageSecretKey          = "KDMP_RSYNC_IMAGE_SECRET"
	RsyncRequestCPU              = "KDMP_RSYNC_REQUEST_CPU"
	RsyncRequestMemory           = "KDMP_RSYNC_REQUEST_MEMORY"
	RsyncLimitCPU                = "KDMP_RSYNC_LIMIT_CPU"
	RsyncLimitMemory             = "KDMP_RSYNC_LIMIT_MEMORY"
	ResticExecutorImageKey       = "KDMP_RESTICEXECUTOR_IMAGE"
	ResticExecutorImageSecretKey = "KDMP_RESTICEXECUTOR_IMAGE_SECRET"
	ResticExecutorRequestCPU     = "KDMP_RESTICEXECUTOR_REQUEST_CPU"
	ResticExecutorRequestMemory  = "KDMP_RESTICEXECUTOR_REQUEST_MEMORY"
	ResticExecutorLimitCPU       = "KDMP_RESTICEXECUTOR_LIMIT_CPU"
	ResticExecutorLimitMemory    = "KDMP_RESTICEXECUTOR_LIMIT_MEMORY"
)

// Default parameters for job options.
const (
	DefaultRsyncRequestCPU             = "1"
	DefaultRsyncRequestMemory          = "700Mi"
	DefaultRsyncLimitCPU               = "2"
	DefaultRsyncLimitMemory            = "1Gi"
	DefaultResticExecutorRequestCPU    = "1"
	DefaultResticExecutorRequestMemory = "700Mi"
	DefaultResticExecutorLimitCPU      = "2"
	DefaultResticExecutorLimitMemory   = "1Gi"
)

// JobState represents a data transfer job state.
type JobState string

const (
	// JobStateInProgress means data transfer is processing.
	JobStateInProgress = "InProgress"
	// JobStateCompleted means data transfer is completed.
	JobStateCompleted = "Completed"
	// JobStateFailed means data transfer is failed.
	JobStateFailed = "Failed"
)

var (
	// ErrJobFailed is a know error for a data transfer job failure.
	ErrJobFailed = fmt.Errorf("data transfer job failed")
)

// Interface defines a data export driver behaviour.
type Interface interface {
	// Name returns a name of the driver.
	Name() string
	// StartJob creates a job for data transfer between volumes.
	StartJob(opts ...JobOption) (id string, err error)
	// DeleteJob stops data transfer between volumes.
	DeleteJob(id string) error
	// JobStatus returns a progress status for a data transfer.
	JobStatus(id string) (status *JobStatus, err error)
}

// JobStatus provides information about data transfer job.
type JobStatus struct {
	ProgressPercents float64
	State            JobState
	Reason           string
}

// IsTransferCompleted allows to check transfer status.
func IsTransferCompleted(progress float64) bool {
	return progress == TransferProgressCompleted
}
