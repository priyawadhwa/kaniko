/*
Copyright 2018 Google LLC

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

package creds

import (
	"context"
	"os"
	"sync"

	"github.com/genuinetools/bpfd/proc"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
)

var (
	setupKeyChainOnce sync.Once
	keyChain          authn.Keychain
)

// GetKeychain returns a keychain for accessing container registries.
func GetKeychain() authn.Keychain {
	setupKeyChainOnce.Do(func() {
		keyChain = authn.NewMultiKeychain(authn.DefaultKeychain)

		// Add the Kubernetes keychain if we're on Kubernetes
		if proc.GetContainerRuntime(0, 0) == proc.RuntimeKubernetes {
			k8sc, err := k8schain.NewNoClient(context.Background())
			if err != nil {
				logrus.Warnf("Error setting up k8schain. Using default keychain %s", err)
				return
			}
			keyChain = authn.NewMultiKeychain(keyChain, k8sc)
		}
	})
	return keyChain
}

func checkDockerConfigJson(keyChain authn.Keychain) (authn.Keychain, error) {
	dockerConfigJSON := dockerConfigJsonLocation()
	// check if the file exists
	contents, err := ioutil.ReadFile(dockerConfigJSON)
	if err != nil {
		return nil, err
	}
	// if it does, parse it into a secret
	var secret v1.Secret
	if err := json.Unmarshal(contents, &secret); err != nil {
		return nil, err
	}

	keyring, err := secrets.MakeDockerKeyring([]v1.Secret{secret}, keyChain)
	if err != nil {
		return nil, err
	}
	return keyring, nil
}

func dockerConfigJsonLocation() string {
	configFile := ".dockerconfigjson"
	if dockerConfig := os.Getenv("DOCKER_CONFIG"); dockerConfig != "" {
		file, err := os.Stat(dockerConfig)
		if err == nil {
			if file.IsDir() {
				return filepath.Join(dockerConfig, configFile)
			}
		} else {
			if os.IsNotExist(err) {
				return string(os.PathSeparator) + filepath.Join("kaniko", ".docker", configFile)
			}
		}
		return filepath.Clean(dockerConfig)
	}
	return string(os.PathSeparator) + filepath.Join("kaniko", ".docker", configFile)
}
