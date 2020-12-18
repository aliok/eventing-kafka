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

package admin

import (
	"context"
	"testing"

	"knative.dev/eventing-kafka/pkg/channel/distributed/controller/env"
	"knative.dev/eventing-kafka/pkg/common/client"

	"knative.dev/pkg/system"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	commontesting "knative.dev/eventing-kafka/pkg/channel/distributed/common/testing"
)

// Mock AdminClient Reference
var mockAdminClient AdminClientInterface

// Test The CreateAdminClient() Kafka Functionality
func TestCreateAdminClientKafka(t *testing.T) {

	// Test Data
	commontesting.SetTestEnvironment(t)
	ctx := context.WithValue(context.TODO(), env.Key{}, &env.Environment{SystemNamespace: system.Namespace()})
	clientId := "TestClientId"
	adminClientType := Kafka
	mockAdminClient = &MockAdminClient{}

	// Replace the NewKafkaAdminClientWrapper To Provide Mock AdminClient & Defer Reset
	NewKafkaAdminClientWrapperRef := NewKafkaAdminClientWrapper
	NewKafkaAdminClientWrapper = func(ctxArg context.Context, saramaConfig *sarama.Config, clientIdArg string, namespaceArg string) (AdminClientInterface, error) {
		assert.Equal(t, ctx, ctxArg)
		assert.Equal(t, clientId, clientIdArg)
		assert.Equal(t, system.Namespace(), namespaceArg)
		assert.Equal(t, adminClientType, adminClientType)
		return mockAdminClient, nil
	}
	defer func() { NewKafkaAdminClientWrapper = NewKafkaAdminClientWrapperRef }()

	saramaConfig, err := client.BuildSaramaConfig(nil, commontesting.SaramaDefaultConfigYaml, nil)
	assert.Nil(t, err)

	// Perform The Test
	adminClient, err := CreateAdminClient(ctx, saramaConfig, clientId, adminClientType)

	// Verify The Results
	assert.Nil(t, err)
	assert.NotNil(t, adminClient)
	assert.Equal(t, mockAdminClient, adminClient)
}

// Test The CreateAdminClient() EventHub Functionality
func TestCreateAdminClientEventHub(t *testing.T) {

	// Test Data
	commontesting.SetTestEnvironment(t)
	ctx := context.WithValue(context.TODO(), env.Key{}, &env.Environment{SystemNamespace: system.Namespace()})
	clientId := "TestClientId"
	adminClientType := EventHub
	mockAdminClient = &MockAdminClient{}

	// Replace the NewEventHubAdminClientWrapper To Provide Mock AdminClient & Defer Reset
	NewEventHubAdminClientWrapperRef := NewEventHubAdminClientWrapper
	NewEventHubAdminClientWrapper = func(ctxArg context.Context, namespaceArg string) (AdminClientInterface, error) {
		assert.Equal(t, ctx, ctxArg)
		assert.Equal(t, system.Namespace(), namespaceArg)
		assert.Equal(t, adminClientType, adminClientType)
		return mockAdminClient, nil
	}
	defer func() { NewEventHubAdminClientWrapper = NewEventHubAdminClientWrapperRef }()

	saramaConfig, err := client.BuildSaramaConfig(nil, commontesting.SaramaDefaultConfigYaml, nil)
	assert.Nil(t, err)

	// Perform The Test
	adminClient, err := CreateAdminClient(ctx, saramaConfig, clientId, adminClientType)

	// Verify The Results
	assert.Nil(t, err)
	assert.NotNil(t, adminClient)
	assert.Equal(t, mockAdminClient, adminClient)
}

// Test The CreateAdminClient Custom Functionality
func TestCreateAdminClientCustom(t *testing.T) {

	// Test Data
	commontesting.SetTestEnvironment(t)
	ctx := context.WithValue(context.TODO(), env.Key{}, &env.Environment{SystemNamespace: system.Namespace()})
	clientId := "TestClientId"
	adminClientType := Custom
	mockAdminClient = &MockAdminClient{}

	// Replace the NewPluginAdminClientWrapper To Provide Mock AdminClient & Defer Reset
	NewCustomAdminClientWrapperRef := NewCustomAdminClientWrapper
	NewCustomAdminClientWrapper = func(ctxArg context.Context, namespaceArg string) (AdminClientInterface, error) {
		assert.Equal(t, ctx, ctxArg)
		assert.Equal(t, system.Namespace(), namespaceArg)
		assert.Equal(t, adminClientType, adminClientType)
		return mockAdminClient, nil
	}
	defer func() { NewCustomAdminClientWrapper = NewCustomAdminClientWrapperRef }()

	saramaConfig, err := client.BuildSaramaConfig(nil, commontesting.SaramaDefaultConfigYaml, nil)
	assert.Nil(t, err)

	// Perform The Test
	adminClient, err := CreateAdminClient(ctx, saramaConfig, clientId, adminClientType)

	// Verify The Results
	assert.Nil(t, err)
	assert.NotNil(t, adminClient)
	assert.Equal(t, mockAdminClient, adminClient)
}

// Test The CreateAdminClient Custom Functionality
func TestCreateAdminClientUnknown(t *testing.T) {

	// Test Data
	ctx := context.TODO()
	clientId := "TestClientId"
	adminClientType := Unknown

	saramaConfig, err := client.BuildSaramaConfig(nil, commontesting.SaramaDefaultConfigYaml, nil)
	assert.Nil(t, err)

	// Perform The Test
	adminClient, err := CreateAdminClient(ctx, saramaConfig, clientId, adminClientType)

	// Verify The Results
	assert.NotNil(t, err)
	assert.Nil(t, adminClient)

}

//
// Mock AdminClient
//

var _ AdminClientInterface = &MockAdminClient{}

type MockAdminClient struct {
	kafkaSecret string
}

func (c MockAdminClient) GetKafkaSecretName(string) string {
	return c.kafkaSecret
}

func (c MockAdminClient) CreateTopic(context.Context, string, *sarama.TopicDetail) *sarama.TopicError {
	return nil
}

func (c MockAdminClient) DeleteTopic(context.Context, string) *sarama.TopicError {
	return nil
}

func (c MockAdminClient) Close() error {
	return nil
}
