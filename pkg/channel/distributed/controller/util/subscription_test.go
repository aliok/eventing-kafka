package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing-kafka/pkg/channel/distributed/controller/constants"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	logtesting "knative.dev/pkg/logging/testing"
)

// Test The SubscriptionLogger Functionality
func TestSubscriptionLogger(t *testing.T) {

	// Test Logger
	logger := logtesting.TestLogger(t).Desugar()

	// Test Data
	subscription := &messagingv1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Name: "TestSubscriptionName", Namespace: "TestSubscriptionNamespace"},
	}

	// Perform The Test
	subscriptionLogger := SubscriptionLogger(logger, subscription)
	assert.NotNil(t, subscriptionLogger)
	assert.NotEqual(t, logger, subscriptionLogger)
	subscriptionLogger.Info("Testing Subscription Logger")
}

// Test The NewSubscriptionControllerRef Functionality
func TestNewSubscriptionControllerRef(t *testing.T) {

	// Test Data
	subscription := &messagingv1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Name: "TestName"},
	}

	// Perform The Test
	controllerRef := NewSubscriptionControllerRef(subscription)

	// Validate Results
	assert.NotNil(t, controllerRef)
	assert.Equal(t, messagingv1.SchemeGroupVersion.String(), controllerRef.APIVersion)
	assert.Equal(t, constants.KnativeSubscriptionKind, controllerRef.Kind)
	assert.Equal(t, subscription.ObjectMeta.Name, controllerRef.Name)
	assert.True(t, *controllerRef.Controller)
}
