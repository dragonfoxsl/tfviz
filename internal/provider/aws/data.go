package aws

import (
	p "github.com/danushkastanley/tfviz/internal/projection"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

// Data adapters (Tier B).
func registerData() {
	register("aws_rds_cluster", Adapter{
		Family: model.FamilyDatabase, Noun: "Aurora cluster",
		LabelFrom: []string{"cluster_identifier"},
		Identity:  []string{"id", "arn", "cluster_identifier"},
		Fields: []p.Field{
			{Key: "cluster_identifier", Label: "Identifier"},
			{Key: "engine", Label: "Engine"},
			{Key: "engine_version", Label: "Engine version"},
			{Key: "storage_encrypted", Label: "Storage encrypted"},
			{Key: "endpoint", Label: "Writer endpoint"},
			{Key: "reader_endpoint", Label: "Reader endpoint"},
			{Key: "master_username", Label: "Master username", Withheld: true},
			{Key: "master_password", Label: "Master password", ChangeOnly: true},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "db_subnet_group_name", Targets: []string{"aws_db_subnet_group"}, Through: true},
			{Type: model.RelSecurityGroupAttachment, Field: "vpc_security_group_ids", Targets: []string{"aws_security_group"}},
			{Type: model.RelEncryptionKey, Field: "kms_key_id", Targets: []string{"aws_kms_key"}},
		},
	})
	register("aws_rds_cluster_instance", Adapter{
		Family: model.FamilyDatabase, Noun: "Aurora instance",
		LabelFrom: []string{"identifier"},
		Placement: Placement{Follow: "cluster_identifier", FollowTargets: []string{"aws_rds_cluster"}},
		Fields: []p.Field{
			{Key: "identifier", Label: "Identifier"},
			{Key: "instance_class", Label: "Instance class"},
			{Key: "availability_zone", Label: "Availability zone"},
			{Key: "writer", Label: "Writer"},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelServiceReference, Field: "cluster_identifier", Targets: []string{"aws_rds_cluster"}},
		},
	})
	register("aws_elasticache_subnet_group", Adapter{
		Family: model.FamilyDatabase, Role: model.RoleAssociation, Noun: "Cache subnet group",
		Identity: []string{"name", "id"},
		Fields:   []p.Field{{Key: "name", Label: "Name"}, {Key: "description", Label: "Description", Withheld: true}},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "subnet_ids", Targets: []string{"aws_subnet"}},
		},
	})
	cache := Adapter{
		Family: model.FamilyDatabase, Noun: "Cache",
		LabelFrom: []string{"replication_group_id", "cluster_id"},
		Fields: []p.Field{
			{Key: "replication_group_id", Label: "Replication group"},
			{Key: "cluster_id", Label: "Cluster"},
			{Key: "engine", Label: "Engine"},
			{Key: "engine_version", Label: "Engine version"},
			{Key: "node_type", Label: "Node type"},
			{Key: "num_cache_clusters", Label: "Nodes"},
			{Key: "transit_encryption_enabled", Label: "Encryption in transit"},
			{Key: "at_rest_encryption_enabled", Label: "Encryption at rest"},
			{Key: "auth_token", Label: "Auth token", ChangeOnly: true},
			{Key: "description", Label: "Description", Withheld: true},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "subnet_group_name", Targets: []string{"aws_elasticache_subnet_group"}, Through: true},
			{Type: model.RelSecurityGroupAttachment, Field: "security_group_ids", Targets: []string{"aws_security_group"}},
			{Type: model.RelEncryptionKey, Field: "kms_key_id", Targets: []string{"aws_kms_key"}},
		},
	}
	register("aws_elasticache_replication_group", cache)
	register("aws_elasticache_cluster", cache)
	register("aws_dynamodb_table", Adapter{
		Family: model.FamilyDatabase, Noun: "DynamoDB table", Placement: regional,
		Identity: []string{"name", "arn", "id"},
		Fields: []p.Field{
			{Key: "name", Label: "Name"},
			{Key: "billing_mode", Label: "Billing mode"},
			{Key: "hash_key", Label: "Partition key"},
			{Key: "range_key", Label: "Sort key"},
			{Key: "server_side_encryption.enabled", Label: "Customer-managed encryption"},
			{Key: "point_in_time_recovery.enabled", Label: "Point-in-time recovery"},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelEncryptionKey, Field: "server_side_encryption.kms_key_arn", Targets: []string{"aws_kms_key"}},
		},
	})
	register("aws_redshift_subnet_group", Adapter{
		Family: model.FamilyDatabase, Role: model.RoleAssociation, Noun: "Redshift subnet group",
		Identity: []string{"name", "id"},
		Fields:   []p.Field{{Key: "name", Label: "Name"}, {Key: "description", Label: "Description", Withheld: true}},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "subnet_ids", Targets: []string{"aws_subnet"}},
		},
	})
	register("aws_redshift_cluster", Adapter{
		Family: model.FamilyDatabase, Noun: "Redshift cluster",
		LabelFrom: []string{"cluster_identifier"},
		Identity:  []string{"id", "arn", "cluster_identifier"},
		Fields: []p.Field{
			{Key: "cluster_identifier", Label: "Identifier"},
			{Key: "node_type", Label: "Node type"},
			{Key: "cluster_type", Label: "Cluster type"},
			{Key: "number_of_nodes", Label: "Nodes"},
			{Key: "encrypted", Label: "Encrypted"},
			{Key: "publicly_accessible", Label: "Publicly accessible"},
			{Key: "master_username", Label: "Master username", Withheld: true},
			{Key: "master_password", Label: "Master password", ChangeOnly: true},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelSubnetMembership, Field: "cluster_subnet_group_name", Targets: []string{"aws_redshift_subnet_group"}, Through: true},
			{Type: model.RelSecurityGroupAttachment, Field: "vpc_security_group_ids", Targets: []string{"aws_security_group"}},
			{Type: model.RelEncryptionKey, Field: "kms_key_id", Targets: []string{"aws_kms_key"}},
		},
	})
	register("aws_ebs_volume", Adapter{
		Family: model.FamilyStorage, Noun: "EBS volume", Placement: regional,
		Fields: []p.Field{
			{Key: "id", Label: "Volume ID"},
			{Key: "availability_zone", Label: "Availability zone"},
			{Key: "size", Label: "Size (GiB)"},
			{Key: "type", Label: "Volume type"},
			{Key: "encrypted", Label: "Encrypted"},
			tagsWithheld,
		},
		Relations: []Relation{
			{Type: model.RelEncryptionKey, Field: "kms_key_id", Targets: []string{"aws_kms_key"}},
		},
	})
	// S3 buckets are regional services; their policies are never exported.
	register("aws_s3_bucket", Adapter{
		Family: model.FamilyStorage, Noun: "S3 bucket", Placement: regional,
		LabelFrom: []string{"bucket"},
		Identity:  []string{"bucket", "id", "arn"},
		Fields: []p.Field{
			{Key: "bucket", Label: "Name"},
			{Key: "arn", Label: "ARN"},
			{Key: "policy", Label: "Bucket policy", Withheld: true},
			tagsWithheld,
		},
	})
}
