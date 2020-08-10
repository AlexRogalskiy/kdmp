package rsync

import (
	"fmt"
	"strings"

	"github.com/portworx/kdmp/pkg/drivers"
	"github.com/portworx/kdmp/pkg/drivers/utils"
	"github.com/portworx/sched-ops/k8s/batch"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Driver is a rsync implementation of the data export interface.
type Driver struct{}

// Name returns a name of the driver.
func (d Driver) Name() string {
	return drivers.Rsync
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

	rsyncJob, err := jobFor(o.SourcePVCName, o.DestinationPVCName, o.Namespace, o.Labels)
	if err != nil {
		return "", err
	}
	if _, err = batch.Instance().CreateJob(rsyncJob); err != nil && !errors.IsAlreadyExists(err) {
		return "", err
	}

	return utils.NamespacedName(rsyncJob.Namespace, rsyncJob.Name), nil
}

// DeleteJob stops data transfer between volumes.
func (d Driver) DeleteJob(id string) error {
	namespace, name, err := utils.ParseJobID(id)
	if err != nil {
		return err
	}

	if err = batch.Instance().DeleteJob(name, namespace); err != nil && !errors.IsNotFound(err) {
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
	if strings.TrimSpace(o.DestinationPVCName) == "" {
		return fmt.Errorf("destination pvc name should be set")
	}
	return nil
}

func jobFor(srcVol, dstVol, namespace string, labels map[string]string) (*batchv1.Job, error) {
	labels = addJobLabels(labels)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      toJobName(srcVol),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:    "rsync",
							Image:   "eeacms/rsync",
							Command: []string{"/bin/sh", "-x", "-c", "ls -la /src; ls -la /dst/; rsync -avz /src/ /dst"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "src-vol",
									MountPath: "/src",
								},
								{
									Name:      "dst-vol",
									MountPath: "/dst",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "src-vol",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: srcVol,
								},
							},
						},
						{
							Name: "dst-vol",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: dstVol,
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
	return fmt.Sprintf("import-rsync-%s", id)
}

func addJobLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[drivers.DriverNameLabel] = drivers.Rsync
	return labels
}
