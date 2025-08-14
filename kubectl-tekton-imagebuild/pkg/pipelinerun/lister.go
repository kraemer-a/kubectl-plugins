package pipelinerun

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonversioned "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/yaml"
)

type Lister struct {
	client tektonversioned.Interface
}

type ListOptions struct {
	Namespace     string
	AllNamespaces bool
	LabelKey      string
	LabelValue    string
	OutputFormat  string
	NoHeaders     bool
}

type PipelineRunInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Status         string            `json:"status"`
	ScheduleValue  string            `json:"scheduleValue"`
	Age            string            `json:"age"`
	StartTime      *metav1.Time      `json:"startTime,omitempty"`
	CompletionTime *metav1.Time      `json:"completionTime,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Pipeline       string            `json:"pipeline,omitempty"`
}

func NewLister(config *rest.Config) (*Lister, error) {
	tektonClient, err := tektonversioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tekton client: %w", err)
	}

	return &Lister{
		client: tektonClient,
	}, nil
}

func (l *Lister) List(opts ListOptions) error {
	ctx := context.Background()

	// Build label selector
	labelSelector := labels.NewSelector()
	if opts.LabelValue != "" {
		requirement, err := labels.NewRequirement(opts.LabelKey, "=", []string{opts.LabelValue})
		if err != nil {
			return fmt.Errorf("failed to create label requirement: %w", err)
		}
		labelSelector = labelSelector.Add(*requirement)
	} else {
		requirement, err := labels.NewRequirement(opts.LabelKey, "exists", []string{})
		if err != nil {
			return fmt.Errorf("failed to create label requirement: %w", err)
		}
		labelSelector = labelSelector.Add(*requirement)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	}

	var pipelineRuns []PipelineRunInfo

	if opts.AllNamespaces {
		prList, err := l.client.TektonV1().PipelineRuns("").List(ctx, listOptions)
		if err != nil {
			return fmt.Errorf("failed to list PipelineRuns: %w", err)
		}
		pipelineRuns = l.extractPipelineRunInfo(prList.Items, opts.LabelKey)
	} else {
		prList, err := l.client.TektonV1().PipelineRuns(opts.Namespace).List(ctx, listOptions)
		if err != nil {
			return fmt.Errorf("failed to list PipelineRuns in namespace %s: %w", opts.Namespace, err)
		}
		pipelineRuns = l.extractPipelineRunInfo(prList.Items, opts.LabelKey)
	}

	// Output results
	switch opts.OutputFormat {
	case "json":
		return l.outputJSON(pipelineRuns)
	case "yaml":
		return l.outputYAML(pipelineRuns)
	case "wide":
		return l.outputTableWide(pipelineRuns, opts.NoHeaders)
	default:
		return l.outputTable(pipelineRuns, opts.NoHeaders)
	}
}

func (l *Lister) extractPipelineRunInfo(items []pipelinev1.PipelineRun, labelKey string) []PipelineRunInfo {
	var results []PipelineRunInfo

	for _, pr := range items {
		info := PipelineRunInfo{
			Name:           pr.Name,
			Namespace:      pr.Namespace,
			Status:         l.getStatus(&pr),
			ScheduleValue:  pr.Labels[labelKey],
			Age:            l.getAge(pr.CreationTimestamp.Time),
			StartTime:      pr.Status.StartTime,
			CompletionTime: pr.Status.CompletionTime,
			Labels:         pr.Labels,
		}

		if pr.Spec.PipelineRef != nil {
			info.Pipeline = pr.Spec.PipelineRef.Name
		}

		results = append(results, info)
	}

	return results
}

func (l *Lister) getStatus(pr *pipelinev1.PipelineRun) string {
	conditions := pr.Status.Conditions
	if len(conditions) == 0 {
		return "Unknown"
	}

	// Find the "Succeeded" condition
	var condition *apis.Condition
	for i := range conditions {
		if string(conditions[i].Type) == "Succeeded" {
			condition = &conditions[i]
			break
		}
	}

	if condition == nil {
		return "Unknown"
	}

	switch condition.Status {
	case corev1.ConditionTrue:
		return "Succeeded"
	case corev1.ConditionFalse:
		reason := condition.Reason
		if reason == "Failed" {
			return "Failed"
		} else if reason == "Cancelled" {
			return "Cancelled"
		} else if reason == "PipelineRunTimeout" {
			return "Timeout"
		}
		return fmt.Sprintf("Failed (%s)", reason)
	case corev1.ConditionUnknown:
		if condition.Reason == "Running" {
			return "Running"
		}
		return "Pending"
	default:
		return "Unknown"
	}
}

func (l *Lister) getAge(creationTime time.Time) string {
	duration := time.Since(creationTime)

	if duration.Hours() > 24*365 {
		years := int(duration.Hours() / (24 * 365))
		return fmt.Sprintf("%dy", years)
	} else if duration.Hours() > 24*30 {
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%dmo", months)
	} else if duration.Hours() > 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	} else if duration.Hours() > 1 {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else if duration.Minutes() > 1 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	return fmt.Sprintf("%ds", int(duration.Seconds()))
}

func (l *Lister) outputTable(pipelineRuns []PipelineRunInfo, noHeaders bool) error {
	table := tablewriter.NewWriter(os.Stdout)

	if !noHeaders {
		table.SetHeader([]string{"NAME", "NAMESPACE", "STATUS", "SCHEDULE", "AGE"})
	}

	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, pr := range pipelineRuns {
		table.Append([]string{
			pr.Name,
			pr.Namespace,
			pr.Status,
			pr.ScheduleValue,
			pr.Age,
		})
	}

	table.Render()
	return nil
}

func (l *Lister) outputTableWide(pipelineRuns []PipelineRunInfo, noHeaders bool) error {
	table := tablewriter.NewWriter(os.Stdout)

	if !noHeaders {
		table.SetHeader([]string{"NAME", "NAMESPACE", "PIPELINE", "STATUS", "SCHEDULE", "START TIME", "COMPLETION TIME", "AGE"})
	}

	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, pr := range pipelineRuns {
		startTime := "-"
		if pr.StartTime != nil {
			startTime = pr.StartTime.Format(time.RFC3339)
		}

		completionTime := "-"
		if pr.CompletionTime != nil {
			completionTime = pr.CompletionTime.Format(time.RFC3339)
		}

		pipeline := pr.Pipeline
		if pipeline == "" {
			pipeline = "-"
		}

		table.Append([]string{
			pr.Name,
			pr.Namespace,
			pipeline,
			pr.Status,
			pr.ScheduleValue,
			startTime,
			completionTime,
			pr.Age,
		})
	}

	table.Render()
	return nil
}

func (l *Lister) outputJSON(pipelineRuns []PipelineRunInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pipelineRuns)
}

func (l *Lister) outputYAML(pipelineRuns []PipelineRunInfo) error {
	data, err := yaml.Marshal(pipelineRuns)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	fmt.Print(string(data))
	return nil
}
