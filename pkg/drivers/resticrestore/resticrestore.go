package resticrestore

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/portworx/kdmp/pkg/drivers"
	"github.com/portworx/kdmp/pkg/drivers/utils"
	kdmpops "github.com/portworx/kdmp/pkg/util/ops"
	"github.com/portworx/sched-ops/k8s/batch"
	coreops "github.com/portworx/sched-ops/k8s/core"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Driver is a resticrestore implementation of the data export interface.
type Driver struct{}

// Name returns a name of the driver.
func (d Driver) Name() string {
	return drivers.ResticRestore
}

// StartJob creates a job for data transfer between volumes.
func (d Driver) StartJob(opts ...drivers.JobOption) (id string, err error) {
	o := drivers.JobOpts{}
	for _, opt := range opts {
		if opt != nil {
			if err := opt(&o); err != nil {
				return "", err
			}
		}
	}

	if err := d.validate(o); err != nil {
		return "", err
	}

	vb, err := kdmpops.Instance().GetVolumeBackup(context.Background(), o.VolumeBackupName, o.VolumeBackupNamespace)
	if err != nil {
		return "", err
	}

	resticSecretName := toJobName(o.SourcePVCName)
	if _, err := coreops.Instance().CreateSecret(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resticSecretName,
			Namespace: o.Namespace,
		},
		StringData: map[string]string{
			drivers.SecretKey: drivers.SecretValue,
		},
	}); err != nil && !apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("create a secret with a restic password: %s", err)
	}

	job, err := jobFor(
		o.Namespace,
		o.DestinationPVCName,
		vb.Spec.BackupLocation.Name,
		vb.Spec.BackupLocation.Namespace,
		vb.Spec.Repository,
		resticSecretName,
		o.Labels)
	if err != nil {
		return "", err
	}
	if _, err = batch.Instance().CreateJob(job); err != nil && !apierrors.IsAlreadyExists(err) {
		return "", err
	}

	return utils.NamespacedName(job.Namespace, job.Name), nil
}

// DeleteJob stops data transfer between volumes.
func (d Driver) DeleteJob(id string) error {
	namespace, name, err := utils.ParseJobID(id)
	if err != nil {
		return err
	}

	if err := coreops.Instance().DeleteSecret(name, namespace); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if err := utils.CleanServiceAccount(name, namespace); err != nil {
		return err
	}

	if err = batch.Instance().DeleteJob(name, namespace); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

// JobStatus returns a progress status for a data transfer.
func (d Driver) JobStatus(id string) (*drivers.JobStatus, error) {
	namespace, name, err := utils.ParseJobID(id)
	if err != nil {
		return utils.ToJobStatus(0, err.Error(), batchv1.JobConditionType("")), nil
	}

	job, err := batch.Instance().GetJob(name, namespace)
	if err != nil {
		return nil, err
	}
	err = utils.JobNodeExists(job)
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch the node info tied to the job %s/%s: %v", namespace, name, err)
		return nil, fmt.Errorf(errMsg)
	}
	var jobStatus batchv1.JobConditionType
	if len(job.Status.Conditions) != 0 {
		jobStatus = job.Status.Conditions[0].Type

	}
	if err != nil {
		errMsg := fmt.Sprintf("failed to rget estart count for job  %s/%s job: %v", namespace, name, err)
		return nil, fmt.Errorf(errMsg)
	}

	if utils.IsJobFailed(job) {
		errMsg := fmt.Sprintf("check %s/%s job for details: %s", namespace, name, drivers.ErrJobFailed)
		return utils.ToJobStatus(0, errMsg, jobStatus), nil
	}
	if !utils.IsJobCompleted(job) {
		// TODO: update progress
		return utils.ToJobStatus(0, "", jobStatus), nil
	}

	return utils.ToJobStatus(drivers.TransferProgressCompleted, "", jobStatus), nil
}

func (d Driver) validate(o drivers.JobOpts) error {
	if o.VolumeBackupName == "" {
		return fmt.Errorf("volumebackup name should be set")
	}
	if o.VolumeBackupNamespace == "" {
		return fmt.Errorf("volumebackup namespace should be set")
	}
	return nil
}

func jobFor(
	namespace,
	pvcName,
	backuplocationName,
	backuplocationNamespace,
	repository,
	secretName string,
	labels map[string]string) (*batchv1.Job, error) {
	labels = addJobLabels(labels)

	resources, err := utils.ResticResourceRequirements()
	if err != nil {
		return nil, err
	}

	genName := toJobName(pvcName)
	if err := utils.SetupServiceAccount(genName, namespace, roleFor()); err != nil {
		return nil, err
	}

	cmd := strings.Join([]string{
		"/resticexecutor",
		"restore",
		"--backup-location",
		backuplocationName,
		"--namespace",
		backuplocationNamespace,
		"--repository",
		repository,
		"--secret-file-path",
		filepath.Join(drivers.SecretMount, drivers.SecretKey),
		"--target-path",
		"/data",
	}, " ")

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      genName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &utils.JobPodBackOffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ImagePullSecrets:   utils.ToImagePullSecret(utils.ResticExecutorImageSecret()),
					ServiceAccountName: genName,
					Containers: []corev1.Container{
						{
							Name:  "resticexecutor",
							Image: utils.ResticExecutorImage(),
							Command: []string{
								"/bin/sh",
								"-x",
								"-c",
								cmd,
							},
							Resources: resources,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "vol",
									MountPath: "/data",
								},
								{
									Name:      "secret",
									MountPath: drivers.SecretMount,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "vol",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
						{
							Name: "secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secretName,
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func toJobName(id string) string {
	return fmt.Sprintf("resticrestore-%s", id)
}

func addJobLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[drivers.DriverNameLabel] = drivers.ResticRestore
	return labels
}

func roleFor() *rbacv1.Role {
	return &rbacv1.Role{
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"stork.libopenstorage.org"},
				Resources: []string{"backuplocations"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"kdmp.portworx.com"},
				Resources: []string{"volumebackups"},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}
}
