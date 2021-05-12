/*
Copyright 2021 cuisongliu@qq.com.

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
package baremetal

import (
	"errors"
	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/infra/aliyun"
	"github.com/alibaba/sealer/logger"
	v1 "github.com/alibaba/sealer/types/api/v1"
	"github.com/alibaba/sealer/utils"
	"strings"
)

type ActionName string

type BaremetalProvider struct {
	Cluster *v1.Cluster
}

const (
	ReconcileInstance ActionName = "ReconcileInstance"
	BindEIP           ActionName = "BindEIP"
)
const (
	Baremetal          = "BAREMETAL"
	Master             = "master"
	Node               = "node"
	BaremetalMasterIPs = common.AliDomain + "MasterIPs"
	BaremetalNodeIPs   = common.AliDomain + "NodeIPs"
)

var RecocileFuncMap = map[ActionName]func(provider *BaremetalProvider) error{
	ReconcileInstance: func(provider *BaremetalProvider) error {
		err := provider.ReconcileIPlist(Master)
		if err != nil {
			return err
		}

		err = provider.ReconcileIPlist(Node)
		if err != nil {
			return err
		}
		return nil
	},
	BindEIP: func(provider *BaremetalProvider) error {
		return provider.BindEipForMaster0()
	},
}

func (a *BaremetalProvider) Reconcile() error {
	if a.Cluster.Annotations == nil {
		a.Cluster.Annotations = make(map[string]string)
	}
	if a.Cluster.DeletionTimestamp != nil {
		logger.Info("DeletionTimestamp not nil Clear Cluster")
		return nil
	}
	todolist := []ActionName{
		ReconcileInstance,
		BindEIP,
	}

	for _, actionname := range todolist {
		err := RecocileFuncMap[actionname](a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *BaremetalProvider) Apply() error {
	return a.Reconcile()
}
func (a *BaremetalProvider) InputIPlist(instanceRole string) (iplist []string, err error) {
	var ipList []string
	var hosts *v1.Hosts
	switch instanceRole {
	case aliyun.Master:
		hosts = &a.Cluster.Spec.Masters
	case aliyun.Node:
		hosts = &a.Cluster.Spec.Nodes
	}
	if hosts == nil {
		return nil, err
	}
	for _, ip := range hosts.IPList {
		ipList = append(ipList, ip)
	}
	return ipList, nil
}
func (a *BaremetalProvider) ReconcileIPlist(instanceRole string) error {
	var hosts *v1.Hosts
	var oldIPList []string
	var oldIPListString string
	switch instanceRole {
	case aliyun.Master:
		hosts = &a.Cluster.Spec.Masters
		if hosts.IPList == nil {
			return errors.New("master IPList not set")
		}
		oldIPListString = a.Cluster.Annotations[BaremetalMasterIPs]
		defer func() {
			a.Cluster.Annotations[BaremetalMasterIPs] = strings.Join(hosts.IPList, ",")
		}()
	case aliyun.Node:
		hosts = &a.Cluster.Spec.Nodes
		if hosts.IPList == nil {
			return nil
		}
		oldIPListString = a.Cluster.Annotations[BaremetalNodeIPs]
		defer func() {
			a.Cluster.Annotations[BaremetalNodeIPs] = strings.Join(hosts.IPList, ",")
		}()
	}
	if hosts == nil {
		return errors.New("hosts not set")
	}
	i := len(hosts.Count)
	oldIPList = strings.Split(oldIPListString, ",")
	ipList, err := a.InputIPlist(instanceRole)
	if err != nil {
		return err
	}
	if len(oldIPList) < i {
		hosts.IPList = utils.AppendIPList(hosts.IPList, ipList)
	} else if len(oldIPList) > i {
		hosts.IPList = utils.ReduceIPList(hosts.IPList, ipList)
	}
	logger.Info("reconcile %s instances success %v ", instanceRole, hosts.IPList)
	return nil
}
func (a *BaremetalProvider) BindEipForMaster0() error {
	masters := a.Cluster.Spec.Masters
	if len(masters.IPList) == 0 {
		return errors.New("can not find master0 ")
	}
	master0 := masters.IPList[0]
	//a.Cluster.Annotations[aliyun.Eip] = "127.0.0.1"
	a.Cluster.Annotations[aliyun.Master0InternalIP] = master0
	return nil
}
