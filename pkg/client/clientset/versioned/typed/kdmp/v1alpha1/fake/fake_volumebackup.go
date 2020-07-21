/*

LICENSE

*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/portworx/kdmp/pkg/apis/kdmp/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVolumeBackups implements VolumeBackupInterface
type FakeVolumeBackups struct {
	Fake *FakeKdmpV1alpha1
	ns   string
}

var volumebackupsResource = schema.GroupVersionResource{Group: "kdmp.portworx.com", Version: "v1alpha1", Resource: "volumebackups"}

var volumebackupsKind = schema.GroupVersionKind{Group: "kdmp.portworx.com", Version: "v1alpha1", Kind: "VolumeBackup"}

// Get takes name of the volumeBackup, and returns the corresponding volumeBackup object, and an error if there is any.
func (c *FakeVolumeBackups) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(volumebackupsResource, c.ns, name), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// List takes label and field selectors, and returns the list of VolumeBackups that match those selectors.
func (c *FakeVolumeBackups) List(opts v1.ListOptions) (result *v1alpha1.VolumeBackupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(volumebackupsResource, volumebackupsKind, c.ns, opts), &v1alpha1.VolumeBackupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VolumeBackupList{ListMeta: obj.(*v1alpha1.VolumeBackupList).ListMeta}
	for _, item := range obj.(*v1alpha1.VolumeBackupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumeBackups.
func (c *FakeVolumeBackups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(volumebackupsResource, c.ns, opts))

}

// Create takes the representation of a volumeBackup and creates it.  Returns the server's representation of the volumeBackup, and an error, if there is any.
func (c *FakeVolumeBackups) Create(volumeBackup *v1alpha1.VolumeBackup) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(volumebackupsResource, c.ns, volumeBackup), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// Update takes the representation of a volumeBackup and updates it. Returns the server's representation of the volumeBackup, and an error, if there is any.
func (c *FakeVolumeBackups) Update(volumeBackup *v1alpha1.VolumeBackup) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(volumebackupsResource, c.ns, volumeBackup), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVolumeBackups) UpdateStatus(volumeBackup *v1alpha1.VolumeBackup) (*v1alpha1.VolumeBackup, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(volumebackupsResource, "status", c.ns, volumeBackup), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// Delete takes name of the volumeBackup and deletes it. Returns an error if one occurs.
func (c *FakeVolumeBackups) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(volumebackupsResource, c.ns, name), &v1alpha1.VolumeBackup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumeBackups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(volumebackupsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VolumeBackupList{})
	return err
}

// Patch applies the patch and returns the patched volumeBackup.
func (c *FakeVolumeBackups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(volumebackupsResource, c.ns, name, pt, data, subresources...), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}
