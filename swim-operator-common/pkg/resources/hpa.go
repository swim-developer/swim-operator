package resources

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HPAParams struct {
	Name      string
	Namespace string
	Labels    map[string]string

	TargetRef  autoscalingv2.CrossVersionObjectReference
	TargetName string

	MinReplicas    *int32
	MaxReplicas    int32
	CPUUtilization *int32

	TargetCPUUtilizationPercentage int32
	ScaleUpStabilization           int32
	ScaleDownStabilization         int32
}

func BuildHPA(p HPAParams) *autoscalingv2.HorizontalPodAutoscaler {
	targetRef := p.TargetRef
	if targetRef.Name == "" && p.TargetName != "" {
		targetRef = autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1", Kind: "Deployment", Name: p.TargetName,
		}
	}

	minReplicas := resolveInt32Ptr(p.MinReplicas, 1)
	maxReplicas := Int32Default(p.MaxReplicas, 5)

	cpuVal := resolveInt32Ptr(p.CPUUtilization, 0)
	if cpuVal == 0 {
		cpuVal = Int32Default(p.TargetCPUUtilizationPercentage, 70)
	}

	scaleUp := Int32Default(p.ScaleUpStabilization, 60)
	scaleDown := Int32Default(p.ScaleDownStabilization, 300)
	policyPeriod := int32(60)
	policyValue := int32(1)

	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: targetRef,
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			Metrics: []autoscalingv2.MetricSpec{{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: &cpuVal,
					},
				},
			}},
			Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: &scaleUp,
					Policies:                   []autoscalingv2.HPAScalingPolicy{{Type: autoscalingv2.PodsScalingPolicy, Value: policyValue, PeriodSeconds: policyPeriod}},
				},
				ScaleDown: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: &scaleDown,
					Policies:                   []autoscalingv2.HPAScalingPolicy{{Type: autoscalingv2.PodsScalingPolicy, Value: policyValue, PeriodSeconds: policyPeriod}},
				},
			},
		},
	}
}

func resolveInt32Ptr(p *int32, def int32) int32 {
	if p != nil && *p > 0 {
		return *p
	}
	return def
}

func BuildSingletonHPA(name, namespace, targetName string, labels map[string]string) *autoscalingv2.HorizontalPodAutoscaler {
	minReplicas := int32(1)
	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace, Labels: labels,
			Annotations: map[string]string{"autoscaling.kubernetes.io/singleton": "true"},
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{APIVersion: "apps/v1", Kind: "Deployment", Name: targetName},
			MinReplicas:    &minReplicas,
			MaxReplicas:    1,
		},
	}
}
