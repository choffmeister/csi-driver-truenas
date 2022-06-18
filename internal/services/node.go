package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/choffmeister/csi-driver-truenas/internal/backends"
	"github.com/choffmeister/csi-driver-truenas/internal/utils"
	proto "github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ proto.NodeServer = (*NodeService)(nil)

type NodeService struct {
	NodeId     string
	mountUtils *utils.MountUtils
	iscsiUtils *utils.ISCSIUtils
}

func NewNodeService(nodeId string) *NodeService {
	return &NodeService{
		NodeId:     nodeId,
		mountUtils: utils.NewMountUtils(),
		iscsiUtils: utils.NewISCSIUtils(),
	}
}

func (s *NodeService) NodeStageVolume(ctx context.Context, req *proto.NodeStageVolumeRequest) (*proto.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: NodeStageVolume")
}

func (s *NodeService) NodeUnstageVolume(ctx context.Context, req *proto.NodeUnstageVolumeRequest) (*proto.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: NodeUnstageVolume")
}

func (s *NodeService) NodePublishVolume(ctx context.Context, req *proto.NodePublishVolumeRequest) (*proto.NodePublishVolumeResponse, error) {
	if ephemeral := req.VolumeContext["csi.storage.k8s.io/ephemeral"]; ephemeral == "true" {
		cifsIP := req.Secrets["cifs-ip"]
		if cifsIP == "" {
			return nil, status.Error(codes.InvalidArgument, "secret value cifs-ip is missing")
		}
		cifsUsername := req.Secrets["cifs-username"]
		if cifsUsername == "" {
			return nil, status.Error(codes.InvalidArgument, "secret value cifs-username is missing")
		}
		cifsPassword := req.Secrets["cifs-password"]
		if cifsPassword == "" {
			return nil, status.Error(codes.InvalidArgument, "secret value cifs-password is missing")
		}
		cifsShare := req.VolumeContext["cifs-share"]
		if cifsShare == "" {
			return nil, status.Error(codes.InvalidArgument, "volume context value cifs-share is missing")
		}
		cifs := fmt.Sprintf("//%s/%s", cifsIP, cifsShare)

		cifsUID := req.VolumeContext["cifs-uid"]
		cifsGID := req.VolumeContext["cifs-gid"]

		options := []string{}
		options = append(options, "username="+cifsUsername)
		options = append(options, "password="+cifsUsername)
		if cifsUID != "" {
			options = append(options, "uid="+cifsUID)
		}
		if cifsGID != "" {
			options = append(options, "gid="+cifsGID)
		}

		utils.Info.Printf("Mounting cifs %s to %s\n", cifs, req.TargetPath)
		if err := os.MkdirAll(req.TargetPath, 0o775); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create mount target path: %v", err))
		}
		_, _, err := utils.Command("mount", "-t", "cifs", "-o", strings.Join(options, ","), cifs, req.TargetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to cifs mount: %v", err))
		}
		return &proto.NodePublishVolumeResponse{}, nil
	}

	iscsiTarget := req.VolumeContext["iscsi-iqn"]
	if iscsiTarget == "" {
		return nil, status.Error(codes.InvalidArgument, "secret value iscsi-iqn is missing")
	}

	backend, err := NewBackendForNodePublish(req.PublishContext, req.Secrets)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create backend: %v", err))
	}
	iscsi, err := backends.LoadISCSISecrets(req.Secrets)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to load iscsi secrets: %v", err))
	}

	podNamespace := req.VolumeContext["csi.storage.k8s.io/pod.namespace"]
	podName := req.VolumeContext["csi.storage.k8s.io/pod.name"]
	if podNamespace != "" && podName != "" {
		err := backend.CommentVolume(ctx, req.VolumeId, fmt.Sprintf("%s/%s", podNamespace, podName))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to set volume comment: %v", err))
		}
	}

	err = s.iscsiUtils.Login(iscsi.PortalIP, iscsi.PortalPort, iscsiTarget)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to log into iscsi session: %v", err))
	}

	// TODO instead of hardcoding a wait time here it should be watched for the device to be available
	time.Sleep(1 * time.Second)
	devicePath, err := s.iscsiUtils.GenerateDeviceName(iscsi.PortalIP, iscsi.PortalPort, iscsiTarget)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to generate iscsi device path: %v", err))
	}

	// TODO get desired file system from request
	if err := s.mountUtils.FormatAndMountDevice(devicePath, req.TargetPath, "ext4"); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to mount device: %v", err))
	}

	utils.Info.Printf("Published volume %s\n", req.VolumeId)
	return &proto.NodePublishVolumeResponse{}, nil
}

func (s *NodeService) NodeUnpublishVolume(ctx context.Context, req *proto.NodeUnpublishVolumeRequest) (*proto.NodeUnpublishVolumeResponse, error) {
	_, err := NewBackendForNodeUnpublish()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create backend: %v", err))
	}
	devicePath, _, err := s.mountUtils.GetDeviceNameFromMount(req.TargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to detect device path from mountpoint: %v", err))
	}
	if err := s.mountUtils.UnmountDevice(req.TargetPath); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to unmount device: %v", err))
	}

	if strings.HasPrefix(devicePath, "//") {
		// looks like cifs
		return &proto.NodeUnpublishVolumeResponse{}, nil
	}

	portalIP, portalPort, iscsiTarget, err := s.iscsiUtils.ParseDeviceName(devicePath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to detect iscsi information from device path: %v", err))
	}

	if err := s.iscsiUtils.Logout(portalIP, portalPort, iscsiTarget); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to log out of iscsi session: %v", err))
	}

	utils.Info.Printf("Unpublished volume %s\n", req.VolumeId)
	return &proto.NodeUnpublishVolumeResponse{}, nil
}

func (s *NodeService) NodeGetVolumeStats(ctx context.Context, req *proto.NodeGetVolumeStatsRequest) (*proto.NodeGetVolumeStatsResponse, error) {
	if req.VolumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "missing volume path")
	}

	totalBytes, usedBytes, availableBytes, err := s.mountUtils.ByteFilesystemStats(req.VolumePath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get volume byte stats: %s", err))
	}

	totalINodes, usedINodes, freeINodes, err := s.mountUtils.INodeFilesystemStats(req.VolumePath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get volume inode stats: %s", err))
	}

	resp := &proto.NodeGetVolumeStatsResponse{
		Usage: []*proto.VolumeUsage{
			{
				Unit:      proto.VolumeUsage_BYTES,
				Available: availableBytes,
				Total:     totalBytes,
				Used:      usedBytes,
			},
			{
				Unit:      proto.VolumeUsage_INODES,
				Available: freeINodes,
				Total:     totalINodes,
				Used:      usedINodes,
			},
		},
	}
	return resp, nil
}

func (s *NodeService) NodeGetCapabilities(ctx context.Context, req *proto.NodeGetCapabilitiesRequest) (*proto.NodeGetCapabilitiesResponse, error) {
	resp := &proto.NodeGetCapabilitiesResponse{
		Capabilities: []*proto.NodeServiceCapability{
			{
				Type: &proto.NodeServiceCapability_Rpc{
					Rpc: &proto.NodeServiceCapability_RPC{
						Type: proto.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &proto.NodeServiceCapability_Rpc{
					Rpc: &proto.NodeServiceCapability_RPC{
						Type: proto.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
	}
	return resp, nil
}

func (s *NodeService) NodeGetInfo(ctx context.Context, req *proto.NodeGetInfoRequest) (*proto.NodeGetInfoResponse, error) {
	resp := &proto.NodeGetInfoResponse{
		NodeId: s.NodeId,
	}
	return resp, nil
}

func (s *NodeService) NodeExpandVolume(ctx context.Context, req *proto.NodeExpandVolumeRequest) (*proto.NodeExpandVolumeResponse, error) {
	size, _, ok := volumeSizeFromCapacityRange(req.GetCapacityRange())
	if !ok {
		return nil, status.Error(codes.OutOfRange, "invalid capacity range")
	}

	_, err := NewBackendForNodeExpandVolume()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create backend: %v", err))
	}

	devicePath, _, err := s.mountUtils.GetDeviceNameFromMount(req.VolumePath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to detect device path from mountpoint: %v", err))
	}
	_, _, iscsiTarget, err := s.iscsiUtils.ParseDeviceName(devicePath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to detect iscsi information from device path: %v", err))
	}
	if err := s.iscsiUtils.Rescan(iscsiTarget); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to rescan iscsi target: %v", err))
	}
	if err := s.mountUtils.ResizeDevice(devicePath, req.VolumePath); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to resize device file system: %v", err))
	}

	utils.Info.Printf("Expanded volume %s\n", req.VolumeId)
	return &proto.NodeExpandVolumeResponse{CapacityBytes: size}, nil
}
