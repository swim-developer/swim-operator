package resources

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func StrDefault(val, def string) string {
	if val != "" {
		return val
	}
	return def
}

func Int32Default(val, def int32) int32 {
	if val > 0 {
		return val
	}
	return def
}

func Int64Default(val, def int64) int64 {
	if val != 0 {
		return val
	}
	return def
}

func ComputeConfigHash(configMap *corev1.ConfigMap, secrets ...*corev1.Secret) string {
	data := make(map[string]interface{})
	if configMap != nil {
		data["configMap"] = configMap.Data
	}
	for i, secret := range secrets {
		if secret != nil {
			data[fmt.Sprintf("secret_%d", i)] = secret.Data
		}
	}
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash[:8])
}

func ResourcesOrDefault(res corev1.ResourceRequirements, reqMem, reqCPU, limMem, limCPU string) corev1.ResourceRequirements {
	if res.Requests == nil && res.Limits == nil {
		return corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(reqMem),
				corev1.ResourceCPU:    resource.MustParse(reqCPU),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(limMem),
				corev1.ResourceCPU:    resource.MustParse(limCPU),
			},
		}
	}
	return res
}

type ProbeOverrides struct {
	InitialDelaySeconds int32
	PeriodSeconds       int32
	TimeoutSeconds      int32
	FailureThreshold    int32
	DefaultInitialDelay int32
	DefaultPeriod       int32
	DefaultTimeout      int32
	DefaultFailure      int32
}

func BuildHTTPProbe(path string, port int, p ProbeOverrides) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: path,
				Port: intstr.FromInt(port),
			},
		},
		InitialDelaySeconds: Int32Default(p.InitialDelaySeconds, p.DefaultInitialDelay),
		PeriodSeconds:       Int32Default(p.PeriodSeconds, p.DefaultPeriod),
		TimeoutSeconds:      Int32Default(p.TimeoutSeconds, p.DefaultTimeout),
		FailureThreshold:    Int32Default(p.FailureThreshold, p.DefaultFailure),
	}
}

func BoolPtr(b bool) *bool { return &b }
