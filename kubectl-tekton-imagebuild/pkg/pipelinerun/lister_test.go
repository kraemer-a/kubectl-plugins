package pipelinerun

import (
	"testing"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name     string
		pr       *pipelinev1.PipelineRun
		expected string
	}{
		{
			name: "Succeeded",
			pr: &pipelinev1.PipelineRun{
				Status: pipelinev1.PipelineRunStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{
							{
								Type:   apis.ConditionType("Succeeded"),
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			expected: "Succeeded",
		},
		{
			name: "Failed",
			pr: &pipelinev1.PipelineRun{
				Status: pipelinev1.PipelineRunStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{
							{
								Type:   apis.ConditionType("Succeeded"),
								Status: corev1.ConditionFalse,
								Reason: "Failed",
							},
						},
					},
				},
			},
			expected: "Failed",
		},
		{
			name: "Running",
			pr: &pipelinev1.PipelineRun{
				Status: pipelinev1.PipelineRunStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{
							{
								Type:   apis.ConditionType("Succeeded"),
								Status: corev1.ConditionUnknown,
								Reason: "Running",
							},
						},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "Cancelled",
			pr: &pipelinev1.PipelineRun{
				Status: pipelinev1.PipelineRunStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{
							{
								Type:   apis.ConditionType("Succeeded"),
								Status: corev1.ConditionFalse,
								Reason: "Cancelled",
							},
						},
					},
				},
			},
			expected: "Cancelled",
		},
		{
			name: "Timeout",
			pr: &pipelinev1.PipelineRun{
				Status: pipelinev1.PipelineRunStatus{
					Status: duckv1.Status{
						Conditions: []apis.Condition{
							{
								Type:   apis.ConditionType("Succeeded"),
								Status: corev1.ConditionFalse,
								Reason: "PipelineRunTimeout",
							},
						},
					},
				},
			},
			expected: "Timeout",
		},
		{
			name: "No conditions",
			pr: &pipelinev1.PipelineRun{
				Status: pipelinev1.PipelineRunStatus{},
			},
			expected: "Unknown",
		},
	}

	lister := &Lister{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lister.getStatus(tt.pr)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Seconds",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "Minutes",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "Hours",
			duration: 3 * time.Hour,
			expected: "3h",
		},
		{
			name:     "Days",
			duration: 2 * 24 * time.Hour,
			expected: "2d",
		},
		{
			name:     "Months",
			duration: 45 * 24 * time.Hour,
			expected: "1mo",
		},
		{
			name:     "Years",
			duration: 400 * 24 * time.Hour,
			expected: "1y",
		},
	}

	lister := &Lister{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creationTime := time.Now().Add(-tt.duration)
			result := lister.getAge(creationTime)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractPipelineRunInfo(t *testing.T) {
	labelKey := "imagebuild.ba.de/imagebuildschedule"
	now := time.Now()

	prs := []pipelinev1.PipelineRun{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pr-1",
				Namespace: "default",
				Labels: map[string]string{
					labelKey: "daily",
				},
				CreationTimestamp: metav1.NewTime(now.Add(-2 * time.Hour)),
			},
			Spec: pipelinev1.PipelineRunSpec{
				PipelineRef: &pipelinev1.PipelineRef{
					Name: "test-pipeline",
				},
			},
			Status: pipelinev1.PipelineRunStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						{
							Type:   apis.ConditionType("Succeeded"),
							Status: corev1.ConditionTrue,
						},
					},
				},
				PipelineRunStatusFields: pipelinev1.PipelineRunStatusFields{
					StartTime:      &metav1.Time{Time: now.Add(-2 * time.Hour)},
					CompletionTime: &metav1.Time{Time: now.Add(-1 * time.Hour)},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pr-2",
				Namespace: "production",
				Labels: map[string]string{
					labelKey: "weekly",
				},
				CreationTimestamp: metav1.NewTime(now.Add(-48 * time.Hour)),
			},
			Status: pipelinev1.PipelineRunStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						{
							Type:   apis.ConditionType("Succeeded"),
							Status: corev1.ConditionUnknown,
							Reason: "Running",
						},
					},
				},
			},
		},
	}

	lister := &Lister{}
	results := lister.extractPipelineRunInfo(prs, labelKey)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if results[0].Name != "test-pr-1" {
		t.Errorf("expected name test-pr-1, got %s", results[0].Name)
	}

	if results[0].ScheduleValue != "daily" {
		t.Errorf("expected schedule value daily, got %s", results[0].ScheduleValue)
	}

	if results[0].Status != "Succeeded" {
		t.Errorf("expected status Succeeded, got %s", results[0].Status)
	}

	if results[0].Pipeline != "test-pipeline" {
		t.Errorf("expected pipeline test-pipeline, got %s", results[0].Pipeline)
	}

	if results[1].Name != "test-pr-2" {
		t.Errorf("expected name test-pr-2, got %s", results[1].Name)
	}

	if results[1].Status != "Running" {
		t.Errorf("expected status Running, got %s", results[1].Status)
	}
}
