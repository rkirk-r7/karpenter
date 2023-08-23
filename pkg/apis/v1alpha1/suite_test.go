/*
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

package v1alpha1_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/mitchellh/hashstructure/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "knative.dev/pkg/logging/testing"
	"knative.dev/pkg/ptr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/karpenter/pkg/apis/v1alpha1"
	"github.com/aws/karpenter/pkg/test"
)

var ctx context.Context

func TestAPIs(t *testing.T) {
	ctx = TestContextWithLogger(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation")
}

var _ = Describe("Validation", func() {
	var ant *v1alpha1.AWSNodeTemplate

	BeforeEach(func() {
		ant = &v1alpha1.AWSNodeTemplate{
			ObjectMeta: metav1.ObjectMeta{Name: strings.ToLower(randomdata.SillyName())},
			Spec: v1alpha1.AWSNodeTemplateSpec{
				AWS: v1alpha1.AWS{
					SubnetSelector:        map[string]string{"foo": "bar"},
					SecurityGroupSelector: map[string]string{"foo": "bar"},
				},
			},
		}
	})

	Context("UserData", func() {
		It("should succeed if user data is empty", func() {
			Expect(ant.Validate(ctx)).To(Succeed())
		})
		It("should fail if launch template is also specified", func() {
			ant.Spec.LaunchTemplateName = ptr.String("someLaunchTemplate")
			ant.Spec.UserData = ptr.String("someUserData")
			Expect(ant.Validate(ctx)).To(Not(Succeed()))
		})
	})
	Context("Tags", func() {
		It("should succeed when tags are empty", func() {
			ant.Spec.Tags = map[string]string{}
			Expect(ant.Validate(ctx)).To(Succeed())
		})
		It("should succeed if tags aren't in restricted tag keys", func() {
			ant.Spec.Tags = map[string]string{
				"karpenter.sh/custom-key": "value",
				"karpenter.sh/managed":    "true",
				"kubernetes.io/role/key":  "value",
			}
			Expect(ant.Validate(ctx)).To(Succeed())
		})
		It("should succeed by validating that regex is properly escaped", func() {
			ant.Spec.Tags = map[string]string{
				"karpenterzsh/provisioner-name": "value",
			}
			Expect(ant.Validate(ctx)).To(Succeed())
			ant.Spec.Tags = map[string]string{
				"kubernetesbio/cluster/test": "value",
			}
			Expect(ant.Validate(ctx)).To(Succeed())
			ant.Spec.Tags = map[string]string{
				"karpenterzsh/managed-by": "test",
			}
			Expect(ant.Validate(ctx)).To(Succeed())
		})
		It("should fail if tags contain a restricted domain key", func() {
			ant.Spec.Tags = map[string]string{
				"karpenter.sh/provisioner-name": "value",
			}
			Expect(ant.Validate(ctx)).To(Not(Succeed()))
			ant.Spec.Tags = map[string]string{
				"kubernetes.io/cluster/test": "value",
			}
			Expect(ant.Validate(ctx)).To(Not(Succeed()))
			ant.Spec.Tags = map[string]string{
				"karpenter.sh/managed-by": "test",
			}
			Expect(ant.Validate(ctx)).To(Not(Succeed()))
		})
	})
	var _ = Describe("AWSNodeTemplate Hash", func() {
		var awsnodetemplatespec v1alpha1.AWSNodeTemplateSpec
		var awsnodetemplate *v1alpha1.AWSNodeTemplate
		BeforeEach(func() {
			awsnodetemplatespec = v1alpha1.AWSNodeTemplateSpec{
				AWS: v1alpha1.AWS{
					AMIFamily:       aws.String(v1alpha1.AMIFamilyAL2),
					Context:         aws.String("context-1"),
					InstanceProfile: aws.String("profile-1"),
					Tags: map[string]string{
						"keyTag-1": "valueTag-1",
						"keyTag-2": "valueTag-2",
					},
					LaunchTemplate: v1alpha1.LaunchTemplate{
						MetadataOptions: &v1alpha1.MetadataOptions{
							HTTPEndpoint: aws.String("test-metadata-1"),
						},
						BlockDeviceMappings: []*v1alpha1.BlockDeviceMapping{
							{
								DeviceName: aws.String("map-device-1"),
							},
							{
								DeviceName: aws.String("map-device-2"),
							},
						},
					},
				},
				UserData:           aws.String("userdata-test-1"),
				DetailedMonitoring: aws.Bool(false),
			}
			awsnodetemplate = test.AWSNodeTemplate(awsnodetemplatespec)
		})
		DescribeTable("should change hash when static fields are updated", func(awsnodetemplatespec v1alpha1.AWSNodeTemplateSpec) {
			expectedHash := awsnodetemplate.Hash()
			updatedAWSNodeTemplate := test.AWSNodeTemplate(*awsnodetemplatespec.DeepCopy(), awsnodetemplatespec)
			actualHash := updatedAWSNodeTemplate.Hash()
			Expect(actualHash).ToNot(Equal(fmt.Sprint(expectedHash)))
		},
			Entry("InstanceProfile Drift", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{InstanceProfile: aws.String("profile-2")}}),
			Entry("UserData Drift", v1alpha1.AWSNodeTemplateSpec{UserData: aws.String("userdata-test-2")}),
			Entry("Tags Drift", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{Tags: map[string]string{"keyTag-test-3": "valueTag-test-3"}}}),
			Entry("MetadataOptions Drift", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{LaunchTemplate: v1alpha1.LaunchTemplate{MetadataOptions: &v1alpha1.MetadataOptions{HTTPEndpoint: aws.String("test-metadata-2")}}}}),
			Entry("BlockDeviceMappings Drift", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{LaunchTemplate: v1alpha1.LaunchTemplate{BlockDeviceMappings: []*v1alpha1.BlockDeviceMapping{{DeviceName: aws.String("map-device-test-3")}}}}}),
			Entry("Context Drift", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{Context: aws.String("context-2")}}),
			Entry("DetailedMonitoring Drift", v1alpha1.AWSNodeTemplateSpec{DetailedMonitoring: aws.Bool(true)}),
			Entry("AMIFamily Drift", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{AMIFamily: aws.String(v1alpha1.AMIFamilyBottlerocket)}}),
			Entry("Reorder Tags", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{Tags: map[string]string{"keyTag-2": "valueTag-2", "keyTag-1": "valueTag-1"}}}),
			Entry("Reorder BlockDeviceMapping", v1alpha1.AWSNodeTemplateSpec{AWS: v1alpha1.AWS{LaunchTemplate: v1alpha1.LaunchTemplate{BlockDeviceMappings: []*v1alpha1.BlockDeviceMapping{{DeviceName: aws.String("map-device-2")}, {DeviceName: aws.String("map-device-1")}}}}}),
		)
		It("should not change hash when behavior/dynamic fields are updated", func() {
			actualHash := awsnodetemplate.Hash()

			expectedHash, err := hashstructure.Hash(awsnodetemplate.Spec, hashstructure.FormatV2, &hashstructure.HashOptions{
				SlicesAsSets:    true,
				IgnoreZeroValue: true,
				ZeroNil:         true,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actualHash).To(Equal(fmt.Sprint(expectedHash)))

			// Update a behavior/dynamic field
			awsnodetemplate.Spec.SubnetSelector = map[string]string{"subnet-test-key": "subnet-test-value"}
			awsnodetemplate.Spec.SecurityGroupSelector = map[string]string{"sg-test-key": "sg-test-value"}
			awsnodetemplate.Spec.AMISelector = map[string]string{"ami-test-key": "ami-test-value"}

			actualHash = awsnodetemplate.Hash()
			Expect(err).ToNot(HaveOccurred())
			Expect(actualHash).To(Equal(fmt.Sprint(expectedHash)))
		})
		It("should expect two provisioner with the same spec to have the same provisioner hash", func() {
			awsnodetemplateTwo := &v1alpha1.AWSNodeTemplate{
				ObjectMeta: metav1.ObjectMeta{Name: strings.ToLower(randomdata.SillyName())},
			}
			awsnodetemplateTwo.Spec = awsnodetemplatespec

			Expect(awsnodetemplate.Hash()).To(Equal(awsnodetemplateTwo.Hash()))
		})
	})
})
