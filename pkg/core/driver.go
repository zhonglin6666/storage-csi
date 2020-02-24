/*
Copyright 2017 The Kubernetes Authors.

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

package core

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
)

type driver struct {
	name       string
	nodeID     string
	endpoint   string
	version    string
	managerURL string

	ids *identityServer
	ns  *nodeServer
	cs  *ControllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

const (
	driverName = "csi-storage"
)

var (
	version = "1.0.0"
)

func NewDriver(nodeID, endpoint, managerURL string) *driver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	d := &driver{
		name:       driverName,
		nodeID:     nodeID,
		endpoint:   endpoint,
		version:    version,
		managerURL: managerURL,
	}

	return d
}

func (d *driver) Run() {
	s := NewNonBlockingGRPCServer()

	d.ns = NewNodeServer(d.name, d.version, d.managerURL)
	d.ids = NewIdentityServer(d.name, d.version)
	d.cs = NewControllerServer(d.nodeID, d.managerURL)

	s.Start(d.endpoint, d.ids, d.cs, d.ns)
	s.Wait()
}
