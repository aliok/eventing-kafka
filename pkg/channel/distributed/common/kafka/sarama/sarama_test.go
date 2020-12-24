/*
Copyright 2020 The Knative Authors

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

package sarama

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"
	commonconfig "knative.dev/eventing-kafka/pkg/channel/distributed/common/config"
	commontesting "knative.dev/eventing-kafka/pkg/channel/distributed/common/testing"
	injectionclient "knative.dev/pkg/client/injection/kube/client"
)

const (
	// EKDefaultConfigYaml is intended to match what's in 200-eventing-kafka-configmap.yaml
	EKDefaultConfigYaml = `
receiver:
  cpuRequest: 100m
  memoryRequest: 50Mi
  replicas: 1
dispatcher:
  cpuRequest: 100m
  memoryRequest: 50Mi
  replicas: 1
kafka:
  topic:
    defaultNumPartitions: 4
    defaultReplicationFactor: 1
    defaultRetentionMillis: 604800000
  adminType: azure
`
	EKDefaultSaramaConfig = `
Net:
  TLS:
    Config:
      ClientAuth: 0
  SASL:
    Mechanism: PLAIN
    Version: 1
Metadata:
  RefreshFrequency: 300000000000
Consumer:
  Offsets:
    AutoCommit:
        Interval: 5000000000
    Retention: 604800000000000
  Return:
    Errors: true
`
)

// Test Enabling Sarama Logging
func TestEnableSaramaLogging(t *testing.T) {

	// Restore Sarama Logger After Test
	saramaLoggerPlaceholder := sarama.Logger
	defer func() {
		sarama.Logger = saramaLoggerPlaceholder
	}()

	// Perform The Test
	EnableSaramaLogging(true)

	// Verify Results (Not Much Is Possible)
	sarama.Logger.Print("TestMessage - Should See")

	EnableSaramaLogging(false)

	// Verify Results Visually
	sarama.Logger.Print("TestMessage - Should Be Hidden")
}

// This test is specifically to validate that our default settings (used in 200-eventing-kafka-configmap.yaml)
// are valid.  If the defaults in the file change, change this test to match for verification purposes.
func TestLoadDefaultSaramaSettings(t *testing.T) {
	commontesting.SetTestEnvironment(t)
	configMap := commontesting.GetTestSaramaConfigMap(EKDefaultSaramaConfig, EKDefaultConfigYaml)
	fakeK8sClient := fake.NewSimpleClientset(configMap)
	ctx := context.WithValue(context.Background(), injectionclient.Key{}, fakeK8sClient)

	config, configuration, err := LoadSettings(ctx, "myClient", nil)
	assert.Nil(t, err)
	// Make sure all of our default Sarama settings were loaded properly
	assert.Equal(t, tls.ClientAuthType(0), config.Net.TLS.Config.ClientAuth)
	assert.Equal(t, sarama.SASLMechanism("PLAIN"), config.Net.SASL.Mechanism)
	assert.Equal(t, int16(1), config.Net.SASL.Version)
	assert.Equal(t, time.Duration(300000000000), config.Metadata.RefreshFrequency)
	assert.Equal(t, time.Duration(5000000000), config.Consumer.Offsets.AutoCommit.Interval)
	assert.Equal(t, time.Duration(604800000000000), config.Consumer.Offsets.Retention)
	assert.Equal(t, true, config.Consumer.Return.Errors)
	assert.Equal(t, "myClient", config.ClientID)

	// Make sure all of our default eventing-kafka settings were loaded properly
	// Specifically checking the type (e.g. int64, int16, int) is important
	assert.Equal(t, resource.Quantity{}, configuration.Receiver.CpuLimit)
	assert.Equal(t, resource.MustParse("100m"), configuration.Receiver.CpuRequest)
	assert.Equal(t, resource.Quantity{}, configuration.Receiver.MemoryLimit)
	assert.Equal(t, resource.MustParse("50Mi"), configuration.Receiver.MemoryRequest)
	assert.Equal(t, 1, configuration.Receiver.Replicas)
	assert.Equal(t, int32(4), configuration.Kafka.Topic.DefaultNumPartitions)
	assert.Equal(t, int16(1), configuration.Kafka.Topic.DefaultReplicationFactor)
	assert.Equal(t, int64(604800000), configuration.Kafka.Topic.DefaultRetentionMillis)
	assert.Equal(t, resource.Quantity{}, configuration.Dispatcher.CpuLimit)
	assert.Equal(t, resource.MustParse("100m"), configuration.Dispatcher.CpuRequest)
	assert.Equal(t, resource.Quantity{}, configuration.Dispatcher.MemoryLimit)
	assert.Equal(t, resource.MustParse("50Mi"), configuration.Dispatcher.MemoryRequest)
	assert.Equal(t, 1, configuration.Dispatcher.Replicas)
	assert.Equal(t, "azure", configuration.Kafka.AdminType)
}

func TestLoadEventingKafkaSettings(t *testing.T) {
	// Set up a configmap and verify that the sarama settings are loaded properly from it
	commontesting.SetTestEnvironment(t)
	configMap := commontesting.GetTestSaramaConfigMap(commontesting.OldSaramaConfig, commontesting.TestEKConfig)
	fakeK8sClient := fake.NewSimpleClientset(configMap)

	ctx := context.WithValue(context.Background(), injectionclient.Key{}, fakeK8sClient)

	saramaConfig, eventingKafkaConfig, err := LoadSettings(ctx, "", nil)
	assert.Nil(t, err)
	verifyTestEKConfigSettings(t, saramaConfig, eventingKafkaConfig)

	// Test the LoadEventingKafkaSettings function by itself
	eventingKafkaConfig, err = LoadEventingKafkaSettings(configMap)
	assert.Nil(t, err)
	assert.Equal(t, commontesting.DispatcherReplicas, fmt.Sprint(eventingKafkaConfig.Dispatcher.Replicas))

	// Verify that invalid YAML returns an error
	configMap.Data[commonconfig.EventingKafkaSettingsConfigKey] = "\tinvalidYAML"
	eventingKafkaConfig, err = LoadEventingKafkaSettings(configMap)
	assert.Nil(t, eventingKafkaConfig)
	assert.NotNil(t, err)

	// Verify that a configmap with no data section returns an error
	configMap.Data = nil
	eventingKafkaConfig, err = LoadEventingKafkaSettings(configMap)
	assert.Nil(t, eventingKafkaConfig)
	assert.NotNil(t, err)

	// Verify that a nil configmap returns an error
	eventingKafkaConfig, err = LoadEventingKafkaSettings(nil)
	assert.Nil(t, eventingKafkaConfig)
	assert.NotNil(t, err)
}

func TestLoadSettings(t *testing.T) {
	// Set up a configmap and verify that the sarama and eventing-kafka settings are loaded properly from it
	ctx := getTestSaramaContext(t, commontesting.OldSaramaConfig, commontesting.TestEKConfig)
	saramaConfig, eventingKafkaConfig, err := LoadSettings(ctx, "", nil)
	assert.Nil(t, err)
	verifyTestEKConfigSettings(t, saramaConfig, eventingKafkaConfig)

	// Verify that a context with no configmap returns an error
	ctx = context.WithValue(context.Background(), injectionclient.Key{}, fake.NewSimpleClientset())
	saramaConfig, eventingKafkaConfig, err = LoadSettings(ctx, "", nil)
	assert.Nil(t, saramaConfig)
	assert.Nil(t, eventingKafkaConfig)
	assert.NotNil(t, err)

	// Verify that a configmap with no data section returns an error
	configMap := commontesting.GetTestSaramaConfigMap("", "")
	configMap.Data = nil
	ctx = context.WithValue(context.Background(), injectionclient.Key{}, fake.NewSimpleClientset(configMap))
	saramaConfig, eventingKafkaConfig, err = LoadSettings(ctx, "", nil)
	assert.Nil(t, saramaConfig)
	assert.Nil(t, eventingKafkaConfig)
	assert.NotNil(t, err)

	// Verify that a configmap with invalid YAML returns an error
	configMap = commontesting.GetTestSaramaConfigMap(commontesting.OldSaramaConfig, "")
	configMap.Data[commontesting.EventingKafkaSettingsConfigKey] = "\tinvalidYaml"
	ctx = context.WithValue(context.Background(), injectionclient.Key{}, fake.NewSimpleClientset(configMap))
	saramaConfig, eventingKafkaConfig, err = LoadSettings(ctx, "", nil)
	assert.Nil(t, saramaConfig)
	assert.Nil(t, eventingKafkaConfig)
	assert.NotNil(t, err)
}

func verifyTestEKConfigSettings(t *testing.T, saramaConfig *sarama.Config, eventingKafkaConfig *commonconfig.EventingKafkaConfig) {
	// Quick checks to make sure the loaded configs aren't complete junk
	assert.Equal(t, commontesting.OldUsername, saramaConfig.Net.SASL.User)
	assert.Equal(t, commontesting.OldPassword, saramaConfig.Net.SASL.Password)
	assert.Equal(t, commontesting.DispatcherReplicas, strconv.Itoa(eventingKafkaConfig.Dispatcher.Replicas))
}

func getTestSaramaContext(t *testing.T, saramaConfig string, eventingKafkaConfig string) context.Context {
	// Set up a configmap and return a context containing that configmap (for tests)
	commontesting.SetTestEnvironment(t)
	configMap := commontesting.GetTestSaramaConfigMap(saramaConfig, eventingKafkaConfig)
	fakeK8sClient := fake.NewSimpleClientset(configMap)
	ctx := context.WithValue(context.Background(), injectionclient.Key{}, fakeK8sClient)
	assert.NotNil(t, ctx)
	return ctx
}
