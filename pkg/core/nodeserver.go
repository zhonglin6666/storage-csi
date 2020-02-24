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
	"fmt"
	"net/http"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"

	"storage-csi/pkg/util"
)

type nodeServer struct {
	nodeID     string
	version    string
	managerURL string
	//mounter mount.Interface
}

func NewNodeServer(nodeID, version, managerURL string) *nodeServer {
	return &nodeServer{
		version:    version,
		managerURL: managerURL,
		nodeID:     nodeID,
	}
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.Infof("zzlin NodePublishVolume begin... req: %#v", req)
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()
	body := makeNodeVolumeRequestBody(targetPath)
	path := fmt.Sprintf("/volumes/%s/mount", volumeID)

	if err := sendRequest(httpMethodPost, ns.managerURL, path, body, nil); err != nil {
		return &csi.NodePublishVolumeResponse{}, err
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func sendRequest(method, baseURL, path string, body interface{}, header map[string]string) error {
	_, code, err := util.Request(method, baseURL, path, body, header)
	glog.Infof("Send request to %s:%s, code: %d, err: %v", baseURL, path, code, err)
	if err != nil || code != http.StatusOK {
		return fmt.Errorf("send request code: %d error: %v", code, err)
	}
	return nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	glog.Infof("zzlin NodeUnpublishVolume begin...")
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()
	body := makeNodeVolumeRequestBody(targetPath)
	path := fmt.Sprintf("/volumes/%s/umount", volumeID)

	if err := sendRequest(httpMethodPost, ns.managerURL, path, body, nil); err != nil {
		return &csi.NodeUnpublishVolumeResponse{}, err
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func makeNodeVolumeRequestBody(targetPath string) map[string]interface{} {
	return map[string]interface{}{
		"targetPath": targetPath,
	}
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return &csi.NodeGetVolumeStatsResponse{}, nil
}

func (ns *nodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (ns *nodeServer) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{}, nil
}
