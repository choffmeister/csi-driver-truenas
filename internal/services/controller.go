package services

import (
	"context"
	"fmt"

	"github.com/choffmeister/csi-driver-truenas/internal/utils"
	proto "github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ proto.ControllerServer = (*ControllerService)(nil)

type ControllerService struct{}

func NewControllerService() *ControllerService {
	return &ControllerService{}
}

func (s *ControllerService) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "missing name")
	}
	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing volume capabilities")
	}
	size, _, ok := volumeSizeFromCapacityRange(req.GetCapacityRange())
	if !ok {
		return nil, status.Error(codes.OutOfRange, "invalid capacity range")
	}
	for i, cap := range req.VolumeCapabilities {
		if !isCapabilitySupported(cap) {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("capability at index %d is not supported", i))
		}
	}

	backend, err := NewBackendForCreateVolume(req.Parameters, req.Secrets)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to create backend: %v", err))
	}
	id, err := backend.CreateVolume(ctx, req.Name, size)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to create volume: %v", err))
	}

	utils.Info.Printf("Created volume %s\n", id)
	resp := &proto.CreateVolumeResponse{
		Volume: &proto.Volume{
			VolumeId:      id,
			CapacityBytes: size,
			VolumeContext: map[string]string{
				// TODO allow to customize format
				"iscsi-iqn": fmt.Sprintf("%s:%s", backend.GetISCSISecrets().BaseIQN, req.Name),
			},
		},
	}
	return resp, nil
}

func (s *ControllerService) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing volume id")
	}

	backend, err := NewBackendForDeleteVolume(req.Secrets)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to create backend: %v", err))
	}
	if err := backend.DeleteVolume(ctx, req.VolumeId); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to delete volume: %v", err))
	}

	utils.Info.Printf("Deleted volume %s\n", req.VolumeId)
	resp := &proto.DeleteVolumeResponse{}
	return resp, nil
}

func (s *ControllerService) ControllerPublishVolume(ctx context.Context, req *proto.ControllerPublishVolumeRequest) (*proto.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: ControllerPublishVolume")
}

func (s *ControllerService) ControllerUnpublishVolume(ctx context.Context, req *proto.ControllerUnpublishVolumeRequest) (*proto.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: ControllerUnpublishVolume")
}

func (s *ControllerService) ControllerExpandVolume(ctx context.Context, req *proto.ControllerExpandVolumeRequest) (*proto.ControllerExpandVolumeResponse, error) {
	size, _, ok := volumeSizeFromCapacityRange(req.GetCapacityRange())
	if !ok {
		return nil, status.Error(codes.OutOfRange, "invalid capacity range")
	}

	backend, err := NewBackendForControllerExpandVolume(req.Secrets)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create backend: %v", err))
	}
	if err := backend.ExpandVolume(ctx, req.VolumeId, size); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to resize device: %v", err))
	}

	utils.Info.Printf("Expanded volume %s\n", req.VolumeId)
	resp := &proto.ControllerExpandVolumeResponse{
		CapacityBytes:         size,
		NodeExpansionRequired: true,
	}
	return resp, nil
}

func (s *ControllerService) ValidateVolumeCapabilities(ctx context.Context, req *proto.ValidateVolumeCapabilitiesRequest) (*proto.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: ValidateVolumeCapabilities")
}

func (s *ControllerService) ListVolumes(ctx context.Context, req *proto.ListVolumesRequest) (*proto.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: ListVolumes")
}

func (s *ControllerService) GetCapacity(ctx context.Context, req *proto.GetCapacityRequest) (*proto.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: GetCapacity")
}

func (s *ControllerService) ControllerGetCapabilities(ctx context.Context, req *proto.ControllerGetCapabilitiesRequest) (*proto.ControllerGetCapabilitiesResponse, error) {
	resp := &proto.ControllerGetCapabilitiesResponse{
		Capabilities: []*proto.ControllerServiceCapability{
			{
				Type: &proto.ControllerServiceCapability_Rpc{
					Rpc: &proto.ControllerServiceCapability_RPC{
						Type: proto.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			// {
			// 	Type: &proto.ControllerServiceCapability_Rpc{
			// 		Rpc: &proto.ControllerServiceCapability_RPC{
			// 			Type: proto.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
			// 		},
			// 	},
			// },
			{
				Type: &proto.ControllerServiceCapability_Rpc{
					Rpc: &proto.ControllerServiceCapability_RPC{
						Type: proto.ControllerServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}
	return resp, nil
}

func (s *ControllerService) ControllerGetVolume(ctx context.Context, req *proto.ControllerGetVolumeRequest) (*proto.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: ControllerGetVolume")
}

func (s *ControllerService) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: CreateSnapshot")
}

func (s *ControllerService) DeleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: DeleteSnapshot")
}

func (s *ControllerService) ListSnapshots(ctx context.Context, req *proto.ListSnapshotsRequest) (*proto.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not supported: ListSnapshots")
}

func volumeSizeFromCapacityRange(cr *proto.CapacityRange) (int64, int64, bool) {
	if cr == nil {
		return DefaultVolumeSize, 0, true
	}

	var minSize int64
	switch {
	case cr.RequiredBytes == 0:
		minSize = DefaultVolumeSize
	case cr.RequiredBytes < 0:
		return 0, 0, false
	default:
		minSize = cr.RequiredBytes
		if minSize < MinVolumeSize {
			minSize = MinVolumeSize
		}
	}

	var maxSize int64
	switch {
	case cr.LimitBytes == 0:
		break // ignore
	case cr.LimitBytes < 0:
		return 0, 0, false
	default:
		maxSize = cr.LimitBytes
	}

	if maxSize != 0 && minSize > maxSize {
		return 0, 0, false
	}

	return minSize, maxSize, true
}

func isCapabilitySupported(cap *proto.VolumeCapability) bool {
	if cap.AccessMode == nil {
		return false
	}
	if cap.AccessMode.Mode != proto.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
		return false
	}
	if cap.GetBlock() != nil {
		return false
	}
	return true
}
