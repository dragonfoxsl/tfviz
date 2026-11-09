package aws

import "maps"

// iconStems names files in the AWS Architecture Icons pack, without the size
// suffix, most specific first. Names were checked against the 31 July 2026
// release. The pack is not distributed with tfviz: users download it and
// pass its folder with --icons. Types with no fitting official icon, such as
// security groups, are left out and keep tfviz's own symbols.
var iconStems = map[string][]string{
	"aws_vpc":                     {"Res_Amazon-VPC_Virtual-private-cloud-VPC", "Arch_Amazon-Virtual-Private-Cloud"},
	"aws_subnet":                  {"Arch_Amazon-Virtual-Private-Cloud"},
	"aws_internet_gateway":        {"Res_Amazon-VPC_Internet-Gateway"},
	"aws_nat_gateway":             {"Res_Amazon-VPC_NAT-Gateway"},
	"aws_eip":                     {"Res_Amazon-EC2_Elastic-IP-Address"},
	"aws_route_table":             {"Res_Amazon-VPC_Router"},
	"aws_route":                   {"Res_Amazon-VPC_Router"},
	"aws_route_table_association": {"Res_Amazon-VPC_Router"},

	"aws_lb":              {"Arch_Elastic-Load-Balancing"},
	"aws_lb_listener":     {"Arch_Elastic-Load-Balancing"},
	"aws_lb_target_group": {"Arch_Elastic-Load-Balancing"},

	"aws_instance":            {"Res_Amazon-EC2_Instance", "Arch_Amazon-EC2"},
	"aws_lambda_function":     {"Res_AWS-Lambda_Lambda-Function", "Arch_AWS-Lambda"},
	"aws_ecs_cluster":         {"Arch_Amazon-Elastic-Container-Service"},
	"aws_ecs_service":         {"Res_Amazon-Elastic-Container-Service_Service", "Arch_Amazon-Elastic-Container-Service"},
	"aws_ecs_task_definition": {"Res_Amazon-Elastic-Container-Service_Task", "Arch_Amazon-Elastic-Container-Service"},
	"aws_eks_cluster":         {"Arch_Amazon-Elastic-Kubernetes-Service"},
	"aws_eks_node_group":      {"Arch_Amazon-Elastic-Kubernetes-Service"},

	"aws_db_instance":                   {"Res_Amazon-Aurora_Amazon-RDS-Instance", "Arch_Amazon-RDS"},
	"aws_rds_cluster":                   {"Arch_Amazon-RDS"},
	"aws_rds_cluster_instance":          {"Res_Amazon-Aurora_Amazon-RDS-Instance", "Arch_Amazon-RDS"},
	"aws_db_subnet_group":               {"Arch_Amazon-RDS"},
	"aws_dynamodb_table":                {"Res_Amazon-DynamoDB_Table", "Arch_Amazon-DynamoDB"},
	"aws_elasticache_cluster":           {"Arch_Amazon-ElastiCache"},
	"aws_elasticache_replication_group": {"Arch_Amazon-ElastiCache"},
	"aws_elasticache_subnet_group":      {"Arch_Amazon-ElastiCache"},
	"aws_s3_bucket":                     {"Res_Amazon-Simple-Storage-Service_Bucket", "Arch_Amazon-Simple-Storage-Service"},

	"aws_msk_cluster":                  {"Arch_Amazon-Managed-Streaming-for-Apache-Kafka"},
	"aws_msk_configuration":            {"Arch_Amazon-Managed-Streaming-for-Apache-Kafka"},
	"aws_msk_scram_secret_association": {"Arch_Amazon-Managed-Streaming-for-Apache-Kafka"},
	"aws_sqs_queue":                    {"Res_Amazon-Simple-Queue-Service_Queue", "Arch_Amazon-Simple-Queue-Service"},
	"aws_sns_topic":                    {"Res_Amazon-Simple-Notification-Service_Topic", "Arch_Amazon-Simple-Notification-Service"},
	"aws_sns_topic_subscription":       {"Arch_Amazon-Simple-Notification-Service"},

	"aws_secretsmanager_secret":         {"Arch_AWS-Secrets-Manager"},
	"aws_secretsmanager_secret_version": {"Arch_AWS-Secrets-Manager"},
	"aws_kms_key":                       {"Arch_AWS-Key-Management-Service"},
	"aws_ssm_parameter":                 {"Res_AWS-Systems-Manager_Parameter-Store", "Arch_AWS-Systems-Manager"},
	"aws_cloudwatch_log_group":          {"Res_Amazon-CloudWatch_Logs", "Arch_Amazon-CloudWatch"},

	// Not yet checked against the pack; an unmatched name keeps tfviz's symbol.
	"aws_vpc_endpoint":          {"Res_Amazon-VPC_Endpoints", "Arch_Amazon-Virtual-Private-Cloud"},
	"aws_ebs_volume":            {"Res_Amazon-Elastic-Block-Store_Volume", "Arch_Amazon-Elastic-Block-Store"},
	"aws_redshift_cluster":      {"Arch_Amazon-Redshift"},
	"aws_redshift_subnet_group": {"Arch_Amazon-Redshift"},
}

// IconStems returns the AWS icon file names to look for, per resource type.
func IconStems() map[string][]string { return maps.Clone(iconStems) }
