package externalsnapshotter

import (
	"github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SnapshotContentOps is an interface to perform k8s VolumeSnapshotContent operations
type SnapshotContentOps interface {
	// CreateSnapshotContent creates the given snapshot content
	CreateSnapshotContent(snap *v1beta1.VolumeSnapshotContent) (*v1beta1.VolumeSnapshotContent, error)
	// GetSnapshotContent returns the snapshot content for given name
	GetSnapshotContent(name string) (*v1beta1.VolumeSnapshotContent, error)
	// ListSnapshotContents lists all snapshot contents
	ListSnapshotContents() (*v1beta1.VolumeSnapshotContentList, error)
	// UpdateSnapshotContent updates the given snapshot content
	UpdateSnapshotContent(snap *v1beta1.VolumeSnapshotContent) (*v1beta1.VolumeSnapshotContent, error)
	// DeleteSnapshotContent deletes the given snapshot content
	DeleteSnapshotContent(name string) error
}

// CreateSnapshotContent creates the given snapshot content.
func (c *Client) CreateSnapshotContent(snap *v1beta1.VolumeSnapshotContent) (*v1beta1.VolumeSnapshotContent, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.client.VolumeSnapshotContents().Create(snap)
}

// GetSnapshotContent returns the snapshot content for given name
func (c *Client) GetSnapshotContent(name string) (*v1beta1.VolumeSnapshotContent, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.client.VolumeSnapshotContents().Get(name, metav1.GetOptions{})
}

// ListSnapshotContents lists all snapshot contents
func (c *Client) ListSnapshotContents() (*v1beta1.VolumeSnapshotContentList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.client.VolumeSnapshotContents().List(metav1.ListOptions{})
}

// UpdateSnapshotContent updates the given snapshot content
func (c *Client) UpdateSnapshotContent(snap *v1beta1.VolumeSnapshotContent) (*v1beta1.VolumeSnapshotContent, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.client.VolumeSnapshotContents().Update(snap)
}

// DeleteSnapshotContent deletes the given snapshot content
func (c *Client) DeleteSnapshotContent(name string) error {
	if err := c.initClient(); err != nil {
		return err
	}
	return c.client.VolumeSnapshotContents().Delete(name, &metav1.DeleteOptions{})
}
