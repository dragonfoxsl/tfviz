package aws

import (
	p "github.com/danushkastanley/tfviz/internal/projection"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

var (
	tagsWithheld = p.Field{Key: "tags", Label: "Tags", Withheld: true}
	vpcScoped    = Placement{VPCField: "vpc_id"}
)

func registerNetwork() {
	register("aws_vpc", Adapter{
		Family: model.FamilyNetwork, Noun: "VPC", DefinesGroup: model.GroupVPC,
		Fields: []p.Field{
			{Key: "id", Label: "VPC ID"},
			{Key: "cidr_block", Label: "IPv4 CIDR"},
			{Key: "enable_dns_hostnames", Label: "DNS hostnames"},
			{Key: "enable_dns_support", Label: "DNS resolution"},
			{Key: "instance_tenancy", Label: "Tenancy"},
			tagsWithheld,
		},
	})
	register("aws_subnet", Adapter{
		Family: model.FamilyNetwork, Noun: "Subnet", DefinesGroup: model.GroupSubnet, Placement: vpcScoped,
		Fields: []p.Field{
			{Key: "id", Label: "Subnet ID"},
			{Key: "cidr_block", Label: "IPv4 CIDR"},
			{Key: "availability_zone", Label: "Availability zone"},
			{Key: "map_public_ip_on_launch", Label: "Public IP on launch"},
			tagsWithheld,
		},
	})
	register("aws_internet_gateway", Adapter{
		Family: model.FamilyNetwork, Noun: "Internet gateway", Placement: vpcScoped,
		Fields: []p.Field{{Key: "id", Label: "Gateway ID"}, tagsWithheld},
	})
	register("aws_eip", Adapter{
		Family: model.FamilyNetwork, Noun: "Elastic IP", Placement: Placement{Regional: true},
		Identity: []string{"id", "allocation_id", "arn"},
		Fields: []p.Field{
			{Key: "public_ip", Label: "Public IP"},
			{Key: "domain", Label: "Domain"},
			tagsWithheld,
		},
	})
	register("aws_nat_gateway", Adapter{
		Family: model.FamilyNetwork, Noun: "NAT gateway", Placement: Placement{SubnetHome: true},
		Fields: []p.Field{
			{Key: "id", Label: "NAT gateway ID"},
			{Key: "subnet_id", Label: "Subnet"},
			{Key: "connectivity_type", Label: "Connectivity"},
			{Key: "public_ip", Label: "Public IP"},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "subnet_id", Targets: []string{"aws_subnet"}},
			{Type: model.RelServiceReference, Field: "allocation_id", Targets: []string{"aws_eip"}},
		},
	})
	register("aws_vpc_endpoint", Adapter{
		Family: model.FamilyNetwork, Noun: "VPC endpoint", Placement: vpcScoped,
		LabelFrom: []string{"tags.Name", "service_name"},
		Fields: []p.Field{
			{Key: "id", Label: "Endpoint ID"},
			{Key: "service_name", Label: "Service"},
			{Key: "vpc_endpoint_type", Label: "Type"},
			{Key: "private_dns_enabled", Label: "Private DNS"},
			{Key: "policy", Label: "Endpoint policy", Withheld: true},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "subnet_ids", Targets: []string{"aws_subnet"}},
			{Type: model.RelSecurityGroupAttachment, Field: "security_group_ids", Targets: []string{"aws_security_group"}},
		},
	})
	register("aws_route_table", Adapter{
		Family: model.FamilyNetwork, Noun: "Route table", Placement: vpcScoped,
		Fields: []p.Field{
			{Key: "id", Label: "Route table ID"},
			{Key: "routes", Label: "Routes", Path: []string{"route"}, Format: formatRoutes},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelRouting, Field: "route.gateway_id", Targets: []string{"aws_internet_gateway"}},
			{Type: model.RelRouting, Field: "route.nat_gateway_id", Targets: []string{"aws_nat_gateway"}},
		},
	})
	register("aws_route_table_association", Adapter{
		Family: model.FamilyNetwork, Role: model.RoleAssociation, Noun: "Route table association",
		Relations: []Relation{
			{Type: model.RelRouteAssociation, From: "route_table_id", FromTargets: []string{"aws_route_table"}, Field: "subnet_id", Targets: []string{"aws_subnet"}},
		},
	})
	register("aws_route", Adapter{
		Family: model.FamilyNetwork, Role: model.RoleAssociation, Noun: "Route",
		Fields: []p.Field{
			{Key: "destination_cidr_block", Label: "Destination"},
		},
		Relations: []Relation{
			{Type: model.RelRouting, From: "route_table_id", FromTargets: []string{"aws_route_table"}, Field: "gateway_id", Targets: []string{"aws_internet_gateway"}},
			{Type: model.RelRouting, From: "route_table_id", FromTargets: []string{"aws_route_table"}, Field: "nat_gateway_id", Targets: []string{"aws_nat_gateway"}},
		},
	})
}
