package labels

import (
	"testing"

	. "github.com/onsi/gomega"
)

const (
	labelKeyApp                 = "app"
	labelKeyName                = "app.kubernetes.io/name"
	labelKeyInstance            = "app.kubernetes.io/instance"
	labelKeyComponent           = "app.kubernetes.io/component"
	labelKeyPartOf              = "app.kubernetes.io/part-of"
	labelKeyManagedBy           = "app.kubernetes.io/managed-by"
	testLabelAppName            = "my-app"
	testLabelComponentPostgres  = "postgres"
	testLabelInstanceMyInstance = "my-instance"
	testLabelOperatorMyOperator = "my-operator"
	testLabelSwimDnotam         = "swim-dnotam"
	testLabelComponentArtemis   = "artemis"
	testLabelInstanceShort      = "inst"
	testLabelMgrShort           = "mgr"
)

func TestStandardLabels_AllKeysPresent(t *testing.T) {
	g := NewWithT(t)
	lbl := StandardLabels(testLabelAppName, testLabelComponentPostgres, testLabelInstanceMyInstance, testLabelOperatorMyOperator)
	g.Expect(lbl).To(HaveLen(6))
	g.Expect(lbl[labelKeyApp]).To(Equal(testLabelAppName))
	g.Expect(lbl[labelKeyName]).To(Equal(testLabelAppName))
	g.Expect(lbl[labelKeyInstance]).To(Equal(testLabelInstanceMyInstance))
	g.Expect(lbl[labelKeyComponent]).To(Equal(testLabelComponentPostgres))
	g.Expect(lbl[labelKeyPartOf]).To(Equal(testLabelOperatorMyOperator))
	g.Expect(lbl[labelKeyManagedBy]).To(Equal(testLabelOperatorMyOperator))
}

func TestStandardLabels_AppEqualsName(t *testing.T) {
	g := NewWithT(t)
	lbl := StandardLabels(testLabelSwimDnotam, testLabelComponentArtemis, testLabelInstanceShort, testLabelMgrShort)
	g.Expect(lbl[labelKeyApp]).To(Equal(lbl[labelKeyName]))
}

func TestStandardLabels_EmptyStringsAllowed(t *testing.T) {
	g := NewWithT(t)
	lbl := StandardLabels("", "", "", "")
	g.Expect(lbl).To(HaveLen(6))
	g.Expect(lbl[labelKeyApp]).To(BeEmpty())
	g.Expect(lbl[labelKeyManagedBy]).To(BeEmpty())
}

func TestStandardLabels_DifferentComponents(t *testing.T) {
	g := NewWithT(t)
	components := []string{testLabelComponentPostgres, testLabelComponentArtemis, "app", "mongodb", "consumer-validator"}
	for _, comp := range components {
		lbl := StandardLabels("name", comp, testLabelInstanceShort, testLabelMgrShort)
		g.Expect(lbl[labelKeyComponent]).To(Equal(comp))
	}
}
