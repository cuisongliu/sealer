// Copyright © 2021 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	v2 "github.com/alibaba/sealer/types/api/v2"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/logger"
	"github.com/alibaba/sealer/pkg/cert"
)

var ErrClusterNotExist = fmt.Errorf("no cluster exist")

func GetDefaultClusterName() (string, error) {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/.sealer", cert.GetUserHomeDir()))
	if err != nil {
		logger.Error(err)
		return "", err
	}
	var clusters []string
	for _, f := range files {
		if f.IsDir() {
			clusters = append(clusters, f.Name())
		}
	}
	if len(clusters) == 1 {
		return clusters[0], nil
	} else if len(clusters) > 1 {
		return "", fmt.Errorf("Select a cluster through the -c parameter: " + strings.Join(clusters, ","))
	}

	return "", ErrClusterNotExist
}

func GetClusterFromFile(filepath string) (cluster *v2.Cluster, err error) {
	cluster = &v2.Cluster{}
	if err = UnmarshalYamlFile(filepath, cluster); err != nil {
		return nil, fmt.Errorf("failed to get cluster from %s, %v", filepath, err)
	}
	cluster.SetAnnotations(common.ClusterfileName, filepath)
	return cluster, nil
}

func GetDefaultCluster() (cluster *v2.Cluster, err error) {
	name, err := GetDefaultClusterName()
	if err != nil {
		return nil, err
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var filepath = fmt.Sprintf("%s/.sealer/%s/Clusterfile", userHome, name)

	return GetClusterFromFile(filepath)
}
