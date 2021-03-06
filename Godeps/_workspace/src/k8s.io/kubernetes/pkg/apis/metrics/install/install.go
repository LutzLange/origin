/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

// Package install installs the metrics API group, making it available as
// an option to all of the API encoding/decoding machinery.
package install

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/latest"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/registered"
	"k8s.io/kubernetes/pkg/api/unversioned"
	_ "k8s.io/kubernetes/pkg/apis/metrics"
	"k8s.io/kubernetes/pkg/apis/metrics/v1alpha1"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/sets"
)

const importPrefix = "k8s.io/kubernetes/pkg/apis/metrics"

var accessor = meta.NewAccessor()

func init() {
	groupMeta, err := latest.RegisterGroup("metrics")
	if err != nil {
		glog.V(4).Infof("%v", err)
		return
	}

	registeredGroupVersions := registered.GroupVersionsForGroup("metrics")
	groupVersion := registeredGroupVersions[0]
	*groupMeta = latest.GroupMeta{
		GroupVersion: groupVersion,
		Codec:        runtime.CodecFor(api.Scheme, groupVersion.String()),
	}

	groupMeta.GroupVersions = registeredGroupVersions

	groupMeta.SelfLinker = runtime.SelfLinker(accessor)

	// the list of kinds that are scoped at the root of the api hierarchy
	// if a kind is not enumerated here, it is assumed to have a namespace scope
	rootScoped := sets.NewString()

	ignoredKinds := sets.NewString()

	groupMeta.RESTMapper = api.NewDefaultRESTMapper(registeredGroupVersions, interfacesFor, importPrefix, ignoredKinds, rootScoped)
	api.RegisterRESTMapper(groupMeta.RESTMapper)
	groupMeta.InterfacesFor = interfacesFor
}

// InterfacesFor returns the default Codec and ResourceVersioner for a given version
// string, or an error if the version is not known.
func interfacesFor(version unversioned.GroupVersion) (*meta.VersionInterfaces, error) {
	switch version {
	case v1alpha1.SchemeGroupVersion:
		return &meta.VersionInterfaces{
			Codec:            v1alpha1.Codec,
			ObjectConvertor:  api.Scheme,
			MetadataAccessor: accessor,
		}, nil
	default:
		g, _ := latest.Group("metrics")
		return nil, fmt.Errorf("unsupported storage version: %s (valid: %v)", version, g.GroupVersions)
	}
}
