package builder_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	api "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/nodebootstrap"
	"github.com/weaveworks/eksctl/pkg/testutils/mockprovider"
	vpcfakes "github.com/weaveworks/eksctl/pkg/vpc/fakes"

	"github.com/weaveworks/eksctl/pkg/cfn/builder"

	"github.com/awslabs/goformation/v4"
	gfneks "github.com/awslabs/goformation/v4/cloudformation/eks"
)

type amiTypeEntry struct {
	nodeGroup *api.ManagedNodeGroup

	expectedAMIType string
}

var _ = DescribeTable("Managed Nodegroup AMI type", func(e amiTypeEntry) {
	clusterConfig := api.NewClusterConfig()
	clusterConfig.Status = &api.ClusterStatus{
		Endpoint: "https://test.com",
	}
	api.SetManagedNodeGroupDefaults(e.nodeGroup, clusterConfig.Metadata, false)
	p := mockprovider.NewMockProvider()
	fakeVPCImporter := new(vpcfakes.FakeImporter)
	bootstrapper, err := nodebootstrap.NewManagedBootstrapper(clusterConfig, e.nodeGroup)
	Expect(err).NotTo(HaveOccurred())
	mockSubnetsAndAZInstanceSupport(clusterConfig, p,
		[]string{"us-west-2a"},
		[]string{}, // local zones
		[]ec2types.InstanceType{
			ec2types.InstanceTypeM5Large,
			ec2types.InstanceTypeP2Xlarge,
			ec2types.InstanceTypeA12xlarge,
			ec2types.InstanceTypeG5gXlarge,
			ec2types.InstanceTypeG4dnXlarge,
		})
	stack := builder.NewManagedNodeGroup(p.EC2(), clusterConfig, e.nodeGroup, nil, bootstrapper, false, fakeVPCImporter)

	Expect(stack.AddAllResources(context.Background())).To(Succeed())
	bytes, err := stack.RenderJSON()
	Expect(err).NotTo(HaveOccurred())

	template, err := goformation.ParseJSON(bytes)
	Expect(err).NotTo(HaveOccurred())
	ngResource, ok := template.Resources["ManagedNodeGroup"]
	Expect(ok).To(BeTrue())
	ng, ok := ngResource.(*gfneks.Nodegroup)
	Expect(ok).To(BeTrue())
	Expect(ng.AmiType.String()).To(Equal(e.expectedAMIType))
},
	Entry("default AMI type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name: "test",
			},
		},
		expectedAMIType: "AL2023_x86_64_STANDARD",
	}),

	Entry("AL2 AMI type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:      "test",
				AMIFamily: api.NodeImageFamilyAmazonLinux2,
			},
		},
		expectedAMIType: "AL2_x86_64",
	}),

	Entry("default Nvidia GPU instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				InstanceType: "p2.xlarge",
			},
		},
		expectedAMIType: "AL2023_x86_64_NVIDIA",
	}),

	Entry("default Neuron GPU instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				InstanceType: "inf1.2xlarge",
			},
		},
		expectedAMIType: "AL2023_x86_64_NEURON",
	}),

	Entry("AL2 GPU instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				AMIFamily:    api.NodeImageFamilyAmazonLinux2,
				InstanceType: "p2.xlarge",
			},
		},
		expectedAMIType: "AL2_x86_64_GPU",
	}),

	Entry("default ARM instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				InstanceType: "a1.2xlarge",
			},
		},
		expectedAMIType: "AL2023_ARM_64_STANDARD",
	}),

	Entry("AL2 ARM instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				AMIFamily:    api.NodeImageFamilyAmazonLinux2,
				InstanceType: "a1.2xlarge",
			},
		},
		expectedAMIType: "AL2_ARM_64",
	}),

	Entry("Bottlerocket AMI type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:      "test",
				AMIFamily: api.NodeImageFamilyBottlerocket,
			},
		},
		expectedAMIType: "BOTTLEROCKET_x86_64",
	}),

	Entry("Bottlerocket on ARM", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				AMIFamily:    api.NodeImageFamilyBottlerocket,
				InstanceType: "a1.2xlarge",
			},
		},
		expectedAMIType: "BOTTLEROCKET_ARM_64",
	}),

	Entry("Bottlerocket on ARM", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				AMIFamily:    api.NodeImageFamilyBottlerocket,
				InstanceType: "a1.2xlarge",
			},
		},
		expectedAMIType: "BOTTLEROCKET_ARM_64",
	}),
	Entry("Bottlerocket ARM GPU instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				AMIFamily:    api.NodeImageFamilyBottlerocket,
				InstanceType: "g5g.xlarge",
			},
		},
		expectedAMIType: "BOTTLEROCKET_ARM_64_NVIDIA",
	}),

	Entry("Bottlerocket x86 Nvidia GPU instance type", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:         "test",
				AMIFamily:    api.NodeImageFamilyBottlerocket,
				InstanceType: "g4dn.xlarge",
			},
		},
		expectedAMIType: "BOTTLEROCKET_x86_64_NVIDIA",
	}),

	Entry("non-native Ubuntu", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:      "test",
				AMIFamily: api.NodeImageFamilyUbuntu2004,
			},
		},
		expectedAMIType: "CUSTOM",
	}),

	Entry("non-native Ubuntu", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:      "test",
				AMIFamily: api.NodeImageFamilyUbuntu2204,
			},
		},
		expectedAMIType: "CUSTOM",
	}),

	Entry("non-native Ubuntu", amiTypeEntry{
		nodeGroup: &api.ManagedNodeGroup{
			NodeGroupBase: &api.NodeGroupBase{
				Name:      "test",
				AMIFamily: api.NodeImageFamilyUbuntuPro2204,
			},
		},
		expectedAMIType: "CUSTOM",
	}),
)
