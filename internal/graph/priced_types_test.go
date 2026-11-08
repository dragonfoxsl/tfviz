package graph

import (
	"strings"
	"testing"

	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

func change(address, attrs string) string {
	typ, _, _ := strings.Cut(address, ".")
	_, name, _ := strings.Cut(address, ".")
	return `{"address":"` + address + `","mode":"managed","type":"` + typ + `","name":"` + name + `","provider_name":"registry.terraform.io/hashicorp/aws",
	 "change":{"actions":["no-op"],"before":` + attrs + `,"after":` + attrs + `,"after_unknown":{},"before_sensitive":{},"after_sensitive":{}}}`
}

// Placement and evidence for types added so that priced resources are not generic cards.
func TestPricedTypesArePlaced(t *testing.T) {
	doc := `{"format_version":"1.2","resource_changes":[` + strings.Join([]string{
		change("aws_vpc.main", `{"id":"vpc-1"}`),
		change("aws_subnet.a", `{"id":"subnet-a","vpc_id":"vpc-1"}`),
		change("aws_subnet.b", `{"id":"subnet-b","vpc_id":"vpc-1"}`),
		change("aws_kms_key.main", `{"id":"key-1","arn":"arn:aws:kms:eu-west-1:111122223333:key/key-1"}`),
		change("aws_redshift_subnet_group.main", `{"id":"wh","name":"wh","subnet_ids":["subnet-a","subnet-b"]}`),
		change("aws_redshift_cluster.main", `{"id":"wh-1","cluster_identifier":"wh-1","cluster_subnet_group_name":"wh"}`),
		change("aws_vpc_endpoint.main", `{"id":"vpce-1","vpc_id":"vpc-1","service_name":"svc","subnet_ids":["subnet-a"]}`),
		change("aws_ebs_volume.main", `{"id":"vol-1","kms_key_id":"arn:aws:kms:eu-west-1:111122223333:key/key-1"}`),
	}, ",") + `]}`
	snap, err := input.Read([]byte(doc), input.KindPlan)
	if err != nil {
		t.Fatal(err)
	}
	g := Build(snap)
	linked := func(from, to string, typ model.RelationshipType) bool {
		for _, r := range relationshipsFrom(t, g, from) {
			if r.target == to && r.typ == typ {
				return true
			}
		}
		return false
	}
	if !linked("aws_redshift_cluster.main", "aws_subnet.b", model.RelSubnetMembership) {
		t.Error("redshift subnets must resolve through its subnet group")
	}
	if !linked("aws_vpc_endpoint.main", "aws_subnet.a", model.RelSubnetMembership) {
		t.Error("endpoint subnet missing")
	}
	if !linked("aws_ebs_volume.main", "aws_kms_key.main", model.RelEncryptionKey) {
		t.Error("volume key missing")
	}
	placed := func(address string) model.Group { return group(g, g.Placement[node(t, g, address).ID].Architecture) }
	if grp := placed("aws_ebs_volume.main"); grp.Kind == model.GroupVPC || grp.Kind == model.GroupSubnet || grp.Kind == model.GroupUnplaced {
		t.Errorf("volume placed in %+v, want outside any VPC", grp)
	}
	if grp := placed("aws_vpc_endpoint.main"); grp.Kind != model.GroupVPC && grp.Kind != model.GroupSubnet {
		t.Errorf("endpoint placed in %+v", grp)
	}
	if grp := placed("aws_redshift_cluster.main"); grp.Kind != model.GroupVPC && grp.Kind != model.GroupSubnet {
		t.Errorf("cluster placed in %+v", grp)
	}
}
