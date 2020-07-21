/*

LICENSE

*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	kdmpv1alpha1 "github.com/portworx/kdmp/pkg/apis/kdmp/v1alpha1"
	versioned "github.com/portworx/kdmp/pkg/client/clientset/versioned"
	internalinterfaces "github.com/portworx/kdmp/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/portworx/kdmp/pkg/client/listers/kdmp/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// VolumeBackupInformer provides access to a shared informer and lister for
// VolumeBackups.
type VolumeBackupInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.VolumeBackupLister
}

type volumeBackupInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewVolumeBackupInformer constructs a new informer for VolumeBackup type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVolumeBackupInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredVolumeBackupInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredVolumeBackupInformer constructs a new informer for VolumeBackup type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredVolumeBackupInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KdmpV1alpha1().VolumeBackups(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KdmpV1alpha1().VolumeBackups(namespace).Watch(options)
			},
		},
		&kdmpv1alpha1.VolumeBackup{},
		resyncPeriod,
		indexers,
	)
}

func (f *volumeBackupInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredVolumeBackupInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *volumeBackupInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kdmpv1alpha1.VolumeBackup{}, f.defaultInformer)
}

func (f *volumeBackupInformer) Lister() v1alpha1.VolumeBackupLister {
	return v1alpha1.NewVolumeBackupLister(f.Informer().GetIndexer())
}
