/*
Copyright 2018 The CDI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
	v1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/forklift/v1beta1"
	scheme "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned/scheme"
)

// OvirtVolumePopulatorsGetter has a method to return a OvirtVolumePopulatorInterface.
// A group's client should implement this interface.
type OvirtVolumePopulatorsGetter interface {
	OvirtVolumePopulators(namespace string) OvirtVolumePopulatorInterface
}

// OvirtVolumePopulatorInterface has methods to work with OvirtVolumePopulator resources.
type OvirtVolumePopulatorInterface interface {
	Create(ctx context.Context, ovirtVolumePopulator *v1beta1.OvirtVolumePopulator, opts v1.CreateOptions) (*v1beta1.OvirtVolumePopulator, error)
	Update(ctx context.Context, ovirtVolumePopulator *v1beta1.OvirtVolumePopulator, opts v1.UpdateOptions) (*v1beta1.OvirtVolumePopulator, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, ovirtVolumePopulator *v1beta1.OvirtVolumePopulator, opts v1.UpdateOptions) (*v1beta1.OvirtVolumePopulator, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.OvirtVolumePopulator, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.OvirtVolumePopulatorList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.OvirtVolumePopulator, err error)
	OvirtVolumePopulatorExpansion
}

// ovirtVolumePopulators implements OvirtVolumePopulatorInterface
type ovirtVolumePopulators struct {
	*gentype.ClientWithList[*v1beta1.OvirtVolumePopulator, *v1beta1.OvirtVolumePopulatorList]
}

// newOvirtVolumePopulators returns a OvirtVolumePopulators
func newOvirtVolumePopulators(c *ForkliftV1beta1Client, namespace string) *ovirtVolumePopulators {
	return &ovirtVolumePopulators{
		gentype.NewClientWithList[*v1beta1.OvirtVolumePopulator, *v1beta1.OvirtVolumePopulatorList](
			"ovirtvolumepopulators",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1beta1.OvirtVolumePopulator { return &v1beta1.OvirtVolumePopulator{} },
			func() *v1beta1.OvirtVolumePopulatorList { return &v1beta1.OvirtVolumePopulatorList{} }),
	}
}
