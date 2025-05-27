package conn

import (
	palaceconfig "github.com/aide-family/moon/pkg/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcd(etcdConf *palaceconfig.Etcd) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:             etcdConf.GetEndpoints(),
		AutoSyncInterval:      etcdConf.GetAutoSyncInterval().AsDuration(),
		DialTimeout:           etcdConf.GetTimeout().AsDuration(),
		DialKeepAliveTime:     etcdConf.GetDialKeepAliveTime().AsDuration(),
		DialKeepAliveTimeout:  etcdConf.GetDialKeepAliveTimeout().AsDuration(),
		MaxCallSendMsgSize:    int(etcdConf.GetMaxCallSendMsgSize()),
		MaxCallRecvMsgSize:    int(etcdConf.GetMaxCallRecvMsgSize()),
		Username:              etcdConf.GetUsername(),
		Password:              etcdConf.GetPassword(),
		RejectOldCluster:      etcdConf.GetRejectOldCluster(),
		PermitWithoutStream:   etcdConf.GetPermitWithoutStream(),
		MaxUnaryRetries:       uint(etcdConf.GetMaxUnaryRetries()),
		BackoffWaitBetween:    etcdConf.GetBackoffWaitBetween().AsDuration(),
		BackoffJitterFraction: etcdConf.GetBackoffJitterFraction(),
	})
}
