package e2e

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"k8s.io/apimachinery/pkg/api/errors"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

func NewClient() (client.Client, error) {
	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("NewClient Error add core scheme: %w", err)
	}

	if err := hivev1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("NewClient Error add HivePolicy scheme: %w", err)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("NewClient Error get kubeconfig: %w", err)
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("NewClient Create client: %w", err)
	}

	return c, nil
}

func CleanHivePolicies(ctx context.Context, c client.Client) error {

	var hivePolicyList hivev1alpha1.HivePolicyList
	if err := c.List(ctx, &hivePolicyList, client.InNamespace(namespaceName)); err != nil {
		return err
	}

	for _, hivePolicy := range hivePolicyList.Items {
		err := c.Delete(ctx, &hivePolicy)
		if err != nil {
			return fmt.Errorf("Error Delte HivePolicy %s: %w", hivePolicy.Name, err)
		}
	}

	return nil
}

func CleanTestPods(ctx context.Context, c client.Client) error {
	
	err := c.Delete(ctx, testPod)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Error Delete Test Pod %s: %w", testPod.Name, err)
	}

	return nil
}


func CreateTestNamespace(ctx context.Context, c client.Client) error {
	
	err := c.Create(ctx, testNamespace)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Error Create test namespace %s: %w", testNamespace.Name, err)
	}

	return nil
}


func DeleteTestNamespace(ctx context.Context, c client.Client) error {
	
	err := c.Delete(ctx, testNamespace)
	if err != nil {
		return fmt.Errorf("Error Delete test namespace %s: %w", testNamespace.Name, err)
	}

	return nil
}
