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

package pos

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"k8s.io/utils/mount"
)

type Driver struct {
	name    string
	nodeID  string
	version string

	endpoint string

	perm *uint32

	//ids *identityServer
	ns    *NodeServer
	cap   map[csi.VolumeCapability_AccessMode_Mode]bool
	cscap []*csi.ControllerServiceCapability
	nscap []*csi.NodeServiceCapability
}

const (
	DriverName = "pos.csi.k8s.io"
	// Address of the POS server
	paramServer = "server"
	// Base directory of the POS server to create volumes under.
	// The base directory must be a direct child of the root directory.
	// The root directory is omitted from the string, for example:
	//     "base" instead of "/base"
	paramShare = "share"
)

var (
	version = "0.0.1"
)

func NewPOSdriver(nodeID, endpoint string, perm *uint32) *Driver {
	glog.Infof("Driver: %v version: %v", DriverName, version)

	n := &Driver{
		name:     DriverName,
		version:  version,
		nodeID:   nodeID,
		endpoint: endpoint,
		cap:      map[csi.VolumeCapability_AccessMode_Mode]bool{},
		perm:     perm,
	}

	vcam := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}
	n.AddVolumeCapabilityAccessModes(vcam)

	// POS plugin does not support ControllerServiceCapability now.
	// If support is added, it should set to appropriate
	// ControllerServiceCapability RPC types.
	n.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})

	n.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_UNKNOWN,
	})
	return n
}

func NewNodeServer(n *Driver, mounter mount.Interface) *NodeServer {
	return &NodeServer{
		Driver:  n,
		mounter: mounter,
	}
}

func (n *Driver) Run(testMode bool) {
	n.ns = NewNodeServer(n, mount.New(""))
	s := NewNonBlockingGRPCServer()
	s.Start(n.endpoint,
		NewDefaultIdentityServer(n),
		// POS plugin has not implemented ControllerServer
		// using default controllerserver.
		NewControllerServer(n),
		n.ns,
		testMode)
	s.Wait()
}

func (n *Driver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability_AccessMode {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
		n.cap[c] = true
	}
	return vca
}

func (n *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability

	for _, c := range cl {
		glog.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}

	n.cscap = csc
}

func (n *Driver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		glog.Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	n.nscap = nsc
}

func IsCorruptedDir(dir string) bool {
	_, pathErr := mount.PathExists(dir)
	fmt.Printf("IsCorruptedDir(%s) returned with error: %v", dir, pathErr)
	return pathErr != nil && mount.IsCorruptedMnt(pathErr)
}
