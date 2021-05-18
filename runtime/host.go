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
package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/logger"
	v1 "github.com/alibaba/sealer/types/api/v1"
	"github.com/alibaba/sealer/utils"
	"github.com/pkg/errors"
)

func (d *Default) hostPreStart(cluster *v1.Cluster) error {
	err := utils.RemoveFileContent(common.EtcHosts, fmt.Sprintf("%s %s", cluster.Spec.Masters.IPList[0], common.APIServerDomain))
	if err != nil {
		return errors.Wrap(err, "remove current cluster old masters to etc hosts failed")
	}
	return nil
}

func (d *Default) hostPostStop(cluster *v1.Cluster) error {
	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		//_, err = utils.CopySingleFile(fmt.Sprintf("/tmp/%s/bin/kubectl", c.ClusterDesired.Name), common.KubectlPath)
		//if err != nil {
		//	logger.Warn("copy kubectl failed: %v", err)
		//	return err
		//}
		//err = utils.Cmd("chmod", "+x", common.KubectlPath)
		//if err != nil {
		//	logger.Warn("chmod kubectl failed: %v", err)
		//	return err
		//}
		err := utils.AppendFile(common.EtcHosts, fmt.Sprintf("%s %s", cluster.Spec.Masters.IPList[0], common.APIServerDomain))
		if err != nil {
			logger.Warn("append desired cluster new masters to etc hosts failed: %v", err)
		}
		_, err = utils.CopySingleFile(fmt.Sprintf(ClusterRootfsWorkspace+"/admin.conf", cluster.Name), common.DefaultKubeconfig)
		if err != nil {
			logger.Warn("copy kube config failed: %v", err)
		}
	} else {
		if err := utils.CleanFiles(common.DefaultKubeconfigDir, common.GetClusterWorkDir(cluster.Name), common.TmpClusterfile, common.KubectlPath); err != nil {
			logger.Warn(err)
		}
		utils.CleanDir(common.GetClusterWorkDir(cluster.Name))
		utils.CleanDir(common.GetClusterRootfsDir(cluster.Name))
	}
	mountClusterDir := filepath.Join(os.TempDir(), cluster.Name)
	utils.CleanDir(mountClusterDir)
	return nil
}
