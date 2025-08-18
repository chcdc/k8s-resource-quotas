package main

import (
	"fmt"
	"strings"
	"context"
	"log"
	"path/filepath"
	"strconv"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	ns     string
	config string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kubectl-resource-quota",
		Short: "View resource quotas for Kubernetes namespaces",
		Long:  `A kubectl plugin to display resource quota information in a formatted table with usage percentages.`,
		RunE:  run,
	}


	rootCmd.Flags().StringVarP(&ns, "namespaces", "n", "", "Namespace(s) to check quotas for. Use comma-separated for multiple: ns1,ns2,ns3 (required)")
	rootCmd.Flags().StringVar(&config, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.MarkFlagRequired("namespaces")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Error:", err)
	}
}

func setupKubernetesClient(configPath string) (*kubernetes.Clientset, error) {
	var kubeConfig string
	if configPath  != "" {
		kubeConfig = configPath
	} else if home := homedir.HomeDir(); home != "" {
		kubeConfig = filepath.Join(home, ".kube", "config")
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error on load Kubeconfig file: %w", err)
	}

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error on create Kubernets Client: %w", err)
	}

	return client, nil
}

func run(cmd *cobra.Command, args []string) error {
	nsList := strings.Split(ns, ",")
	var emptyNs []string

	client, err := setupKubernetesClient(config)
	if err != nil {
		return err
	}

	for _, namespace := range nsList {
		namespace = strings.TrimSpace(namespace)
		if namespace == "" {
			continue
		}

		fmt.Printf("Checking namespace: %s\n", namespace)

		quotaList, err := client.CoreV1().ResourceQuotas(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("ERROR getting quotas for %s: %v\n", namespace, err)
			continue
		}

		if len(quotaList.Items) == 0 {
			emptyNs = append(emptyNs, namespace)
			fmt.Printf("No quotas found in %s\n", namespace)
			continue
		}

		for _, quota := range quotaList.Items {
			// print header
			fmt.Println("==========================================")
			fmt.Printf("Name:\t\t%s\n", quota.Name)
			fmt.Printf("Namespace:\t%s\n", quota.Namespace)
			fmt.Println("==========================================")
			fmt.Printf("Resource\t\tUsed\t\tHard\t\tPercentage\n")
			fmt.Printf("--------\t\t----\t\t----\t\t----------\n")

			for resourceName, hardLimit := range quota.Status.Hard {
				used := quota.Status.Used[resourceName]

				usedStr := used.String()
				hardStr := hardLimit.String()

				var percentage float64
				if !hardLimit.IsZero() {
					usedVal := used.AsApproximateFloat64()
					hardVal := hardLimit.AsApproximateFloat64()
					if hardVal > 0 {
						percentage = (usedVal / hardVal) * 100
					}
				}


				if strings.HasSuffix(usedStr, "Ki") || strings.HasSuffix(usedStr, "Mi") || strings.HasSuffix(usedStr, "Gi") {
				} else {
					if val, err := strconv.ParseInt(usedStr, 10, 64); err == nil && val > 1024 {
						if val >= 1024*1024*1024 {
							usedStr = fmt.Sprintf("%.2fGi", float64(val)/(1024*1024*1024))
						} else if val >= 1024*1024 {
							usedStr = fmt.Sprintf("%.2fMi", float64(val)/(1024*1024))
						}
					}
				}

				if strings.HasSuffix(hardStr, "Ki") || strings.HasSuffix(hardStr, "Mi") || strings.HasSuffix(hardStr, "Gi") {
				} else {
					if val, err := strconv.ParseInt(hardStr, 10, 64); err == nil && val > 1024 {
						if val >= 1024*1024*1024 {
							hardStr = fmt.Sprintf("%.2fGi", float64(val)/(1024*1024*1024))
						} else if val >= 1024*1024 {
							hardStr = fmt.Sprintf("%.2fMi", float64(val)/(1024*1024))
						}
					}
				}

				fmt.Printf("%s\t\t%s\t\t%s\t\t%.1f%%\n", string(resourceName), usedStr, hardStr, percentage)
			}
			fmt.Println()
		}
	}

	if len(emptyNs) > 0 {
		fmt.Printf("Namespaces with no quotas: %s\n", strings.Join(emptyNs, ", "))
	}

	return nil
}