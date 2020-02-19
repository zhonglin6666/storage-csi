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
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	oneGB = 1073741824

	mountPath = "/persistentvolumes"
)

type ControllerServer struct {
	nodeID string
	caps   []*csi.ControllerServiceCapability
}

type nfsVolume struct {
	VolName            string `json:"volName"`
	VolID              string `json:"volID"`
	ArchiveOnDelete    string `json:"archiveOnDelete"`
	Provisioner        string `json:"provisioner"`
	VolSize            int64  `json:"volSize"`
	AdminID            string `json:"adminId"`
	UserID             string `json:"userId"`
	Mounter            string `json:"mounter"`
	DisableInUseChecks bool   `json:"disableInUseChecks"`
	ClusterID          string `json:"clusterId"`
}

func NewControllerServer(nodeID string) *ControllerServer {
	return &ControllerServer{
		nodeID: nodeID,
		caps: addControllerServiceCapabilities(
			[]csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
				csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
				csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
				csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
			}),
	}
}

func (cs *ControllerServer) validateVolumeReq(req *csi.CreateVolumeRequest) error {
	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Infof("invalid create volume req: %v", protosanitizer.StripSecrets(req))
		return err
	}
	// Check sanity of request Name, Volume Capabilities
	if len(req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "Volume Name cannot be empty")
	}
	if req.VolumeCapabilities == nil {
		return status.Error(codes.InvalidArgument, "Volume Capabilities cannot be empty")
	}
	return nil
}

func (cs *ControllerServer) validateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range cs.caps {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return status.Errorf(codes.InvalidArgument, "unsupported capability %s", c)
}

func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	glog.Infof("controller server create volume begin request: %v", req)
	if err := cs.validateVolumeReq(req); err != nil {
		return nil, err
	}

	//util.VolumeNameMutex.LockKey(req.GetName())
	//defer func() {
	//	if err := util.VolumeNameMutex.UnlockKey(req.GetName()); err != nil {
	//		glog.Warningf("failed to unlock mutex volume:%s %v", req.GetName(), err)
	//	}
	//}()

	nfsVol, err := parseVolCreateRequest(req)
	if err != nil {
		return nil, err
	}

	volumeContext := req.GetParameters()

	//if _, ok := volumeContext["share"]; ok {
	//	volumeContext["share"] = fmt.Sprintf("%s/%s", nfsVol.Share, nfsVol.VolID)
	//}
	glog.Infof("create volume success, volumeContext: %v", volumeContext)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      nfsVol.VolID,
			CapacityBytes: nfsVol.VolSize,
			VolumeContext: req.GetParameters(),
		},
	}, nil
}

func parseVolCreateRequest(req *csi.CreateVolumeRequest) (*nfsVolume, error) {
	nfsVol, err := getnfsVolumeOptions(req.GetParameters(), true)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Generating Volume Name and Volume ID, as according to CSI spec they MUST be different
	nfsVol.VolName = req.GetName()
	volumeID := "csi-nfs-vol-" + uuid.NewUUID().String()
	nfsVol.VolID = volumeID
	// Volume Size - Default is 1 GiB
	volSizeBytes := int64(oneGB)
	if req.GetCapacityRange() != nil {
		volSizeBytes = req.GetCapacityRange().GetRequiredBytes()
	}

	nfsVol.VolSize = volSizeBytes

	return nfsVol, nil
}

func getnfsVolumeOptions(volOptions map[string]string, disableInUseChecks bool) (*nfsVolume, error) {
	//var (
	//	ok bool
	//)

	nfsVol := &nfsVolume{}
	//nfsVol.Server, ok = volOptions["server"]
	//if !ok {
	//	return nil, errors.New("missing required parameter pool")
	//}
	//
	//nfsVol.Share, ok = volOptions["share"]
	//if !ok {
	//	return nil, errors.New("missing required parameter share")
	//}
	//
	//nfsVol.ArchiveOnDelete, ok = volOptions["volOptions"]
	//if !ok {
	//	nfsVol.ArchiveOnDelete = "false"
	//}

	return nfsVol, nil
}

// DeleteVolume deletes the volume in backend
func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	//glog.Infof("DeleteVolume req: %v", req.VolumeId)
	//if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
	//	glog.Warningf("invalid delete volume req: %v", protosanitizer.StripSecrets(req))
	//	return nil, err
	//}
	//// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeID := req.GetVolumeId()
	//util.VolumeNameMutex.LockKey(volumeID)
	//
	//if volumeID == "" {
	//	return nil, errors.New("volume id is nil")
	//}
	//
	//defer func() {
	//	if err := util.VolumeNameMutex.UnlockKey(volumeID); err != nil {
	//		glog.Warningf("failed to unlock mutex volume:%s %v", volumeID, err)
	//	}
	//}()

	nfsVol := &nfsVolume{}
	volName := nfsVol.VolName
	fullPath := filepath.Join(mountPath, volumeID)

	//mounter := mount.New("")
	//err := mounter.Mount(fmt.Sprintf("%v:%v", nfsVol.Server, nfsVol.Share), mountPath, "nfs", nil)
	//if err != nil {
	//	if os.IsPermission(err) {
	//		return nil, status.Error(codes.PermissionDenied, err.Error())
	//	}
	//	if strings.Contains(err.Error(), "invalid argument") {
	//		return nil, status.Error(codes.InvalidArgument, err.Error())
	//	}
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//defer func() {
	//	err = mount.CleanupMountPoint(mountPath, mounter, false)
	//	if err != nil {
	//		glog.Errorf("umount %v error: %v", mountPath, err)
	//	}
	//}()

	glog.Infof("deleting volume %s path: %v", volName, fullPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		glog.Warningf("path %s does not exist, deletion skipped", fullPath)
		return nil, nil
	}

	if err := os.RemoveAll(fullPath); err != nil {
		glog.Error("nfs volume can not remove path: %v", fullPath)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (cs *ControllerServer) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}

func (cs *ControllerServer) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return &csi.ListVolumesResponse{}, nil
}

func (cs *ControllerServer) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return &csi.GetCapacityResponse{}, nil
}

func (cs *ControllerServer) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.caps,
	}, nil
}

func (cs *ControllerServer) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return &csi.CreateSnapshotResponse{}, nil
}

func (cs *ControllerServer) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return &csi.DeleteSnapshotResponse{}, nil
}

func (cs *ControllerServer) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return &csi.ListSnapshotsResponse{}, nil
}

func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return &csi.ControllerExpandVolumeResponse{}, nil
}

func addControllerServiceCapabilities(caps []csi.ControllerServiceCapability_RPC_Type) []*csi.ControllerServiceCapability {
	var cscs []*csi.ControllerServiceCapability

	for _, cap := range caps {
		glog.Infof("Enabling controller service capablity: %s", cap.String())
		cscs = append(cscs, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}
	return cscs
}
