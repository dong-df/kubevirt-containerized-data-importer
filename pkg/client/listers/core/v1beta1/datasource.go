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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
	v1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

// DataSourceLister helps list DataSources.
// All objects returned here must be treated as read-only.
type DataSourceLister interface {
	// List lists all DataSources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.DataSource, err error)
	// DataSources returns an object that can list and get DataSources.
	DataSources(namespace string) DataSourceNamespaceLister
	DataSourceListerExpansion
}

// dataSourceLister implements the DataSourceLister interface.
type dataSourceLister struct {
	listers.ResourceIndexer[*v1beta1.DataSource]
}

// NewDataSourceLister returns a new DataSourceLister.
func NewDataSourceLister(indexer cache.Indexer) DataSourceLister {
	return &dataSourceLister{listers.New[*v1beta1.DataSource](indexer, v1beta1.Resource("datasource"))}
}

// DataSources returns an object that can list and get DataSources.
func (s *dataSourceLister) DataSources(namespace string) DataSourceNamespaceLister {
	return dataSourceNamespaceLister{listers.NewNamespaced[*v1beta1.DataSource](s.ResourceIndexer, namespace)}
}

// DataSourceNamespaceLister helps list and get DataSources.
// All objects returned here must be treated as read-only.
type DataSourceNamespaceLister interface {
	// List lists all DataSources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.DataSource, err error)
	// Get retrieves the DataSource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.DataSource, error)
	DataSourceNamespaceListerExpansion
}

// dataSourceNamespaceLister implements the DataSourceNamespaceLister
// interface.
type dataSourceNamespaceLister struct {
	listers.ResourceIndexer[*v1beta1.DataSource]
}
