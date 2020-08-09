package ops

import (
	"fmt"
	"time"

	kdmpv1alpha1 "github.com/portworx/kdmp/pkg/apis/kdmp/v1alpha1"
	"github.com/portworx/sched-ops/k8s/errors"
	"github.com/portworx/sched-ops/task"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// VolumeBackupOps is an interface to perform k8s VolumeBackup operations
type VolumeBackupOps interface {
	// CreateVolumeBackup creates the VolumeBackup
	CreateVolumeBackup(*kdmpv1alpha1.VolumeBackup) (*kdmpv1alpha1.VolumeBackup, error)
	// GetVolumeBackup gets the VolumeBackup
	GetVolumeBackup(string, string) (*kdmpv1alpha1.VolumeBackup, error)
	// ListVolumeBackups lists all the VolumeBackups
	ListVolumeBackups(string) (*kdmpv1alpha1.VolumeBackupList, error)
	// UpdateVolumeBackup updates the VolumeBackup
	UpdateVolumeBackup(*kdmpv1alpha1.VolumeBackup) (*kdmpv1alpha1.VolumeBackup, error)
	// DeleteVolumeBackup deletes the VolumeBackup
	DeleteVolumeBackup(string, string) error
	// ValidateVolumeBackup validates the VolumeBackup
	ValidateVolumeBackup(string, string, time.Duration, time.Duration) error
}

// CreateVolumeBackup creates the VolumeBackup
func (c *Client) CreateVolumeBackup(backupLocation *kdmpv1alpha1.VolumeBackup) (*kdmpv1alpha1.VolumeBackup, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kdmp.KdmpV1alpha1().VolumeBackups(backupLocation.Namespace).Create(backupLocation)
}

// GetVolumeBackup gets the VolumeBackup
func (c *Client) GetVolumeBackup(name string, namespace string) (*kdmpv1alpha1.VolumeBackup, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kdmp.KdmpV1alpha1().VolumeBackups(namespace).Get(name, metav1.GetOptions{})
}

// ListVolumeBackups lists all the VolumeBackups
func (c *Client) ListVolumeBackups(namespace string) (*kdmpv1alpha1.VolumeBackupList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kdmp.KdmpV1alpha1().VolumeBackups(namespace).List(metav1.ListOptions{})
}

// UpdateVolumeBackup updates the VolumeBackup
func (c *Client) UpdateVolumeBackup(backupLocation *kdmpv1alpha1.VolumeBackup) (*kdmpv1alpha1.VolumeBackup, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kdmp.KdmpV1alpha1().VolumeBackups(backupLocation.Namespace).Update(backupLocation)
}

// PatchVolumeBackup applies a patch for a given volumebackup.
func (c *Client) PatchVolumeBackup(name, ns string, pt types.PatchType, jsonPatch []byte) (*kdmpv1alpha1.VolumeBackup, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kdmp.KdmpV1alpha1().VolumeBackups(ns).Patch(name, pt, jsonPatch)
}

// DeleteVolumeBackup deletes the VolumeBackup
func (c *Client) DeleteVolumeBackup(name string, namespace string) error {
	if err := c.initClient(); err != nil {
		return err
	}
	return c.kdmp.KdmpV1alpha1().VolumeBackups(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deleteForegroundPolicy,
	})
}

// ValidateVolumeBackup validates the VolumeBackup
func (c *Client) ValidateVolumeBackup(name, namespace string, timeout, retryInterval time.Duration) error {
	if err := c.initClient(); err != nil {
		return err
	}
	t := func() (interface{}, bool, error) {
		resp, err := c.GetVolumeBackup(name, namespace)
		if err != nil {
			return "", true, &errors.ErrFailedToValidateCustomSpec{
				Name:  name,
				Cause: fmt.Sprintf("VolumeBackup failed. Error: %v", err),
				Type:  resp,
			}
		}
		return "", false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, timeout, retryInterval); err != nil {
		return err
	}
	return nil
}
