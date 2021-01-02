package main

import (
	"fmt"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/ec2"
	"github.com/awslabs/goformation/v4/cloudformation/ecs"
	"github.com/awslabs/goformation/v4/cloudformation/elasticloadbalancingv2"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
	"github.com/awslabs/goformation/v4/cloudformation/route53"
)

func main() {
	template := cloudformation.NewTemplate()

	template.Parameters["VpcId"] = cloudformation.Parameter{
		Type: "AWS::EC2::VPC::Id",
	}

	template.Parameters["SubnetId1"] = cloudformation.Parameter{
		Type: "AWS::EC2::Subnet::Id",
	}

	template.Parameters["SubnetId2"] = cloudformation.Parameter{
		Type: "AWS::EC2::Subnet::Id",
	}

	template.Parameters["SecurityGroupId"] = cloudformation.Parameter{
		Type: "AWS::EC2::SecurityGroup::Id",
	}

	template.Parameters["CertificateArn"] = cloudformation.Parameter{
		Type: "String",
	}

	template.Resources["MyCluster"] = &ecs.Cluster{
		ClusterName: "go-current-tempature",
	}

	template.Resources["MyDefinition"] = &ecs.TaskDefinition{
		Family: "go-current-tempature",
		ContainerDefinitions: []ecs.TaskDefinition_ContainerDefinition{
			ecs.TaskDefinition_ContainerDefinition{
				Name: "go-current-tempature",
				Image: cloudformation.Join("", []string{
					cloudformation.Ref("AWS::AccountId"),
					".dkr.ecr.",
					cloudformation.Ref("AWS::Region"),
					".amazonaws.com/go-current-tempature",
				}),
				PortMappings: []ecs.TaskDefinition_PortMapping{
					ecs.TaskDefinition_PortMapping{
						ContainerPort: 8000,
					},
				},
				Secrets: []ecs.TaskDefinition_Secret{
					ecs.TaskDefinition_Secret{
						Name: "OWM_API_KEY",
						ValueFrom: cloudformation.Join(":", []string{
							"arn:aws:secretsmanager",
							cloudformation.Ref("AWS::Region"),
							cloudformation.Ref("AWS::AccountId"),
							"secret:go-current-tempature/owmApiKey",
						}),
					},
				},
			},
		},
		RequiresCompatibilities: []string{"FARGATE"},
		NetworkMode:             "awsvpc",
		Cpu:                     "256",
		Memory:                  "512",
		ExecutionRoleArn:        cloudformation.Ref("MyRole"),
	}

	template.Resources["MyRole"] = &iam.Role{
		RoleName: "go-current-tempature",
		AssumeRolePolicyDocument: map[string]interface{}{
			"Statement": map[string]interface{}{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Service": []string{
						"ecs-tasks.amazonaws.com",
					},
				},
				"Action": []string{
					"sts:AssumeRole",
				},
			},
		},
		ManagedPolicyArns: []string{
			"arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
		},
		Policies: []iam.Role_Policy{
			iam.Role_Policy{
				PolicyName: "go-current-tempature",
				PolicyDocument: map[string]interface{}{
					"Statement": map[string]interface{}{
						"Effect": "Allow",
						"Action": []string{
							"secretsmanager:GetSecretValue",
						},
						"Resource": []string{
							cloudformation.Join(":", []string{
								"arn:aws:secretsmanager",
								cloudformation.Ref("AWS::Region"),
								cloudformation.Ref("AWS::AccountId"),
								"secret:go-current-tempature/owmApiKey*",
							}),
						},
					},
				},
			},
		},
	}

	template.Resources["MyAlb"] = &elasticloadbalancingv2.LoadBalancer{
		Name: "go-current-tempature",
		Subnets: []string{
			cloudformation.Ref("SubnetId1"),
			cloudformation.Ref("SubnetId2"),
		},
	}

	template.Resources["MyHttpsListener"] = &elasticloadbalancingv2.Listener{
		Port:     443,
		Protocol: "HTTPS",
		DefaultActions: []elasticloadbalancingv2.Listener_Action{
			elasticloadbalancingv2.Listener_Action{
				Type:           "forward",
				TargetGroupArn: cloudformation.Ref("MyTg"),
			},
		},
		LoadBalancerArn: cloudformation.Ref("MyAlb"),
		Certificates: []elasticloadbalancingv2.Listener_Certificate{
			elasticloadbalancingv2.Listener_Certificate{
				CertificateArn: cloudformation.Ref("CertificateArn"),
			},
		},
	}

	template.Resources["MyHttpListener"] = &elasticloadbalancingv2.Listener{
		Port:     80,
		Protocol: "HTTP",
		DefaultActions: []elasticloadbalancingv2.Listener_Action{
			elasticloadbalancingv2.Listener_Action{
				Type: "redirect",
				RedirectConfig: &elasticloadbalancingv2.Listener_RedirectConfig{
					Protocol:   "HTTPS",
					Port:       "443",
					StatusCode: "HTTP_301",
				},
			},
		},
		LoadBalancerArn: cloudformation.Ref("MyAlb"),
	}

	template.Resources["MyTg"] = &elasticloadbalancingv2.TargetGroup{
		Name:                       "go-current-tempature",
		Protocol:                   "HTTP",
		Port:                       8000,
		VpcId:                      cloudformation.Ref("VpcId"),
		AWSCloudFormationDependsOn: []string{"MyAlb"},
		TargetType:                 "ip",
	}

	template.Resources["MySg"] = &ec2.SecurityGroup{
		GroupName:        "go-current-tempature",
		GroupDescription: "Allow ingress 8000",
		SecurityGroupIngress: []ec2.SecurityGroup_Ingress{
			ec2.SecurityGroup_Ingress{
				IpProtocol:            "tcp",
				FromPort:              8000,
				ToPort:                8000,
				SourceSecurityGroupId: cloudformation.Ref("SecurityGroupId"),
			},
		},
	}

	template.Resources["MyService"] = &ecs.Service{
		ServiceName:                "go-current-tempature",
		AWSCloudFormationDependsOn: []string{"MyHttpsListener"},
		Cluster:                    cloudformation.Ref("MyCluster"),
		LoadBalancers: []ecs.Service_LoadBalancer{
			ecs.Service_LoadBalancer{
				TargetGroupArn: cloudformation.Ref("MyTg"),
				ContainerName:  "go-current-tempature",
				ContainerPort:  8000,
			},
		},
		TaskDefinition: cloudformation.Ref("MyDefinition"),
		NetworkConfiguration: &ecs.Service_NetworkConfiguration{
			AwsvpcConfiguration: &ecs.Service_AwsVpcConfiguration{
				AssignPublicIp: "ENABLED",
				SecurityGroups: []string{
					cloudformation.GetAtt("MySg", "GroupId"),
				},
				Subnets: []string{
					cloudformation.Ref("SubnetId1"),
				},
			},
		},
		LaunchType: "FARGATE",
	}

	template.Resources["MyRecordSet"] = &route53.RecordSet{
		HostedZoneName: "whs.hk.",
		Name:           "t.whs.hk",
		Type:           "A",
		AliasTarget: &route53.RecordSet_AliasTarget{
			DNSName:      cloudformation.GetAtt("MyAlb", "DNSName"),
			HostedZoneId: cloudformation.GetAtt("MyAlb", "CanonicalHostedZoneID"),
		},
	}

	j, err := template.JSON()
	if err != nil {
		fmt.Printf("Failed to generate JSON: %s\n", err)
	} else {
		fmt.Printf("%s\n", string(j))
	}
}
