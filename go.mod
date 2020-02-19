module storage-csi

go 1.13

require (
	github.com/container-storage-interface/spec v1.2.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/kubernetes-csi/csi-lib-utils v0.7.0
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/zhonglin6666/kube-nfs-csi v0.0.0-20190606070822-458b4c8de4d8
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	google.golang.org/grpc v1.27.1
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.0.0-20191015221719-7d47edc353ef
	k8s.io/apimachinery v0.17.1-beta.0
	k8s.io/apiserver v0.0.0-20191015220424-a5d070e3855f // indirect
	k8s.io/client-go v0.17.0
	k8s.io/cloud-provider v0.0.0-20191004111010-9775d7be8494 // indirect
	k8s.io/kubernetes v1.14.8
	k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
)
