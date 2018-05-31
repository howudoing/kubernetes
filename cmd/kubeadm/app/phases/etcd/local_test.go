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

package etcd

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	testutil "k8s.io/kubernetes/cmd/kubeadm/test"
)

func TestGetEtcdPodSpec(t *testing.T) {

	// Creates a Master Configuration
	cfg := &kubeadmapi.MasterConfiguration{
		KubernetesVersion: "v1.7.0",
		Etcd: kubeadmapi.Etcd{
			Local: &kubeadmapi.LocalEtcd{
				DataDir: "/var/lib/etcd",
				Image:   "",
			},
		},
	}

	// Executes GetEtcdPodSpec
	spec := GetEtcdPodSpec(cfg)

	// Assert each specs refers to the right pod
	if spec.Spec.Containers[0].Name != kubeadmconstants.Etcd {
		t.Errorf("getKubeConfigSpecs spec for etcd contains pod %s, expects %s", spec.Spec.Containers[0].Name, kubeadmconstants.Etcd)
	}
}

func TestCreateLocalEtcdStaticPodManifestFile(t *testing.T) {

	// Create temp folder for the test case
	tmpdir := testutil.SetupTempDir(t)
	defer os.RemoveAll(tmpdir)

	// Creates a Master Configuration
	cfg := &kubeadmapi.MasterConfiguration{
		KubernetesVersion: "v1.7.0",
		Etcd: kubeadmapi.Etcd{
			Local: &kubeadmapi.LocalEtcd{
				DataDir: "/var/lib/etcd",
				Image:   "k8s.gcr.io/etcd",
			},
		},
	}

	// Execute createStaticPodFunction
	manifestPath := filepath.Join(tmpdir, kubeadmconstants.ManifestsSubDirName)
	err := CreateLocalEtcdStaticPodManifestFile(manifestPath, cfg)
	if err != nil {
		t.Errorf("Error executing CreateEtcdStaticPodManifestFile: %v", err)
	}

	// Assert expected files are there
	testutil.AssertFilesCount(t, manifestPath, 1)
	testutil.AssertFileExists(t, manifestPath, kubeadmconstants.Etcd+".yaml")
}

func TestGetEtcdCommand(t *testing.T) {
	var tests = []struct {
		cfg      *kubeadmapi.MasterConfiguration
		expected []string
	}{
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Etcd: kubeadmapi.Etcd{Local: &kubeadmapi.LocalEtcd{DataDir: "/var/lib/etcd"}},
			},
			expected: []string{
				"etcd",
				"--listen-client-urls=https://127.0.0.1:2379",
				"--advertise-client-urls=https://127.0.0.1:2379",
				"--data-dir=/var/lib/etcd",
				"--cert-file=" + kubeadmconstants.EtcdServerCertName,
				"--key-file=" + kubeadmconstants.EtcdServerKeyName,
				"--trusted-ca-file=" + kubeadmconstants.EtcdCACertName,
				"--client-cert-auth=true",
				"--peer-cert-file=" + kubeadmconstants.EtcdPeerCertName,
				"--peer-key-file=" + kubeadmconstants.EtcdPeerKeyName,
				"--peer-trusted-ca-file=" + kubeadmconstants.EtcdCACertName,
				"--snapshot-count=10000",
				"--peer-client-cert-auth=true",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Etcd: kubeadmapi.Etcd{
					Local: &kubeadmapi.LocalEtcd{
						DataDir: "/var/lib/etcd",
						ExtraArgs: map[string]string{
							"listen-client-urls":    "https://10.0.1.10:2379",
							"advertise-client-urls": "https://10.0.1.10:2379",
						},
					},
				},
			},
			expected: []string{
				"etcd",
				"--listen-client-urls=https://10.0.1.10:2379",
				"--advertise-client-urls=https://10.0.1.10:2379",
				"--data-dir=/var/lib/etcd",
				"--cert-file=" + kubeadmconstants.EtcdServerCertName,
				"--key-file=" + kubeadmconstants.EtcdServerKeyName,
				"--trusted-ca-file=" + kubeadmconstants.EtcdCACertName,
				"--client-cert-auth=true",
				"--peer-cert-file=" + kubeadmconstants.EtcdPeerCertName,
				"--peer-key-file=" + kubeadmconstants.EtcdPeerKeyName,
				"--peer-trusted-ca-file=" + kubeadmconstants.EtcdCACertName,
				"--snapshot-count=10000",
				"--peer-client-cert-auth=true",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Etcd: kubeadmapi.Etcd{Local: &kubeadmapi.LocalEtcd{DataDir: "/etc/foo"}},
			},
			expected: []string{
				"etcd",
				"--listen-client-urls=https://127.0.0.1:2379",
				"--advertise-client-urls=https://127.0.0.1:2379",
				"--data-dir=/etc/foo",
				"--cert-file=" + kubeadmconstants.EtcdServerCertName,
				"--key-file=" + kubeadmconstants.EtcdServerKeyName,
				"--trusted-ca-file=" + kubeadmconstants.EtcdCACertName,
				"--client-cert-auth=true",
				"--peer-cert-file=" + kubeadmconstants.EtcdPeerCertName,
				"--peer-key-file=" + kubeadmconstants.EtcdPeerKeyName,
				"--peer-trusted-ca-file=" + kubeadmconstants.EtcdCACertName,
				"--snapshot-count=10000",
				"--peer-client-cert-auth=true",
			},
		},
	}

	for _, rt := range tests {
		actual := getEtcdCommand(rt.cfg)
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getEtcdCommand:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}
