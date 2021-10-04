package kopiarestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/portworx/kdmp/pkg/apis/kdmp/v1alpha1"
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

// Driver is a kopiarestore implementation of the data export interface.
type Driver struct{}

// Name returns a name of the driver.
func (d Driver) Name() string {
	return drivers.KopiaRestore
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

	jobName := o.DataExportName
	job, err := jobFor(
		o,
		vb,
		jobName,
		utils.FrameCredSecretName(utils.RestoreJobPrefix, o.DataExportName),
	)
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
		return utils.ToJobStatus(0, err.Error()), nil
	}

	job, err := batch.Instance().GetJob(name, namespace)
	if err != nil {
		return nil, err
	}
	if utils.IsJobFailed(job) {
		errMsg := fmt.Sprintf("check %s/%s job for details: %s", namespace, name, drivers.ErrJobFailed)
		return utils.ToJobStatus(0, errMsg), nil
	}
	if !utils.IsJobCompleted(job) {
		// TODO: update progress
		return utils.ToJobStatus(0, ""), nil
	}

	return utils.ToJobStatus(drivers.TransferProgressCompleted, ""), nil
}

func (d Driver) validate(o drivers.JobOpts) error {
	if o.DestinationPVCName == "" {
		return fmt.Errorf("destination pvc name should be set")
	}
	if o.VolumeBackupName == "" {
		return fmt.Errorf("volumebackup name should be set")
	}
	if o.VolumeBackupNamespace == "" {
		return fmt.Errorf("volumebackup namespace should be set")
	}
	return nil
}

func jobFor(
	jobOption drivers.JobOpts,
	vb *v1alpha1.VolumeBackup,
	jobName,
	credSecretName string,
) (*batchv1.Job, error) {
	labels := addJobLabels(jobOption.Labels)

	resources, err := utils.JobResourceRequirements()
	if err != nil {
		return nil, err
	}

	if err := utils.SetupServiceAccount(jobName, jobOption.Namespace, roleFor()); err != nil {
		return nil, err
	}

	cmd := strings.Join([]string{
		"/kopiaexecutor",
		"restore",
		"--volume-backup-name",
		jobOption.VolumeBackupName,
		"--backup-location",
		vb.Spec.BackupLocation.Name,
		"--backup-location-namespace",
		vb.Spec.BackupLocation.Namespace,
		"--repository",
		vb.Spec.Repository,
		"--restore-namespace",
		jobOption.Namespace,
		"--credentials",
		credSecretName,
		"--target-path",
		"/data",
		"--snapshot-id",
		vb.Status.SnapshotID,
	}, " ")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: jobOption.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ImagePullSecrets:   utils.ToImagePullSecret(utils.KopiaExecutorImageSecret()),
					ServiceAccountName: jobName,
					Containers: []corev1.Container{
						{
							Name:            "kopiaexecutor",
							Image:           utils.KopiaExecutorImage(),
							ImagePullPolicy: corev1.PullAlways,
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
									Name:      "cred-secret",
									MountPath: drivers.KopiaCredSecretMount,
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
									ClaimName: jobOption.DestinationPVCName,
								},
							},
						},
						{
							Name: "cred-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: credSecretName,
								},
							},
						},
					},
				},
			},
		},
	}

	if drivers.CertFilePath != "" {
		volumeMount := corev1.VolumeMount{
			Name:      "tls-secret",
			MountPath: drivers.CertMount,
			ReadOnly:  true,
		}

		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(
			job.Spec.Template.Spec.Containers[0].VolumeMounts,
			volumeMount,
		)

		volume := corev1.Volume{
			Name: "tls-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: jobOption.CertSecretName,
				},
			},
		}

		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)

		env := []corev1.EnvVar{
			{
				Name:  drivers.CertDirPath,
				Value: drivers.CertMount,
			},
		}

		job.Spec.Template.Spec.Containers[0].Env = env
	}

	return job, nil
}

func addJobLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[drivers.DriverNameLabel] = drivers.KopiaRestore
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
