# Supported resource types

<!-- Generated from internal/provider/aws. Regenerate with: go test ./internal/provider/aws -update -->

Supported means the type's placement, approved details and relationships are interpreted and tested against Terraform 1.16 and OpenTofu 1.13 fixtures with AWS provider 6.67. It does not mean every provider attribute is shown: tags, descriptions, policies, secret contents and similar free-form or sensitive values are withheld by design.

Any other type appears with limited detail and no placement, and `--strict` treats it as a failure.

## Networking

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_eip` | Resource | Tags |
| `aws_internet_gateway` | Resource | Tags |
| `aws_nat_gateway` | Resource | Tags |
| `aws_route` | Connection between resources | — |
| `aws_route_table` | Resource | Tags |
| `aws_route_table_association` | Connection between resources | — |
| `aws_subnet` | Container (subnet) | Tags |
| `aws_vpc` | Container (vpc) | Tags |
| `aws_vpc_endpoint` | Resource | Endpoint policy, Tags |

## Security groups

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_security_group` | Resource | Description, Tags |
| `aws_security_group_rule` | Connection between resources | Description |
| `aws_vpc_security_group_egress_rule` | Connection between resources | Description |
| `aws_vpc_security_group_ingress_rule` | Connection between resources | Description |

## Load balancing

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_lb` | Resource | Tags |
| `aws_lb_listener` | Resource | — |
| `aws_lb_target_group` | Resource | Tags |

## Compute

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_ecs_cluster` | Resource | Tags |
| `aws_ecs_service` | Resource | Tags |
| `aws_ecs_task_definition` | Resource | Container definitions, Tags |
| `aws_eks_cluster` | Resource | Certificate authority, Tags |
| `aws_eks_node_group` | Resource | Tags |
| `aws_instance` | Resource | User data, Tags |
| `aws_lambda_function` | Resource | Environment variables, Tags |

## Databases and caches

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_db_instance` | Resource | Master username, Master password (change only), Tags |
| `aws_db_subnet_group` | Connection between resources | Description |
| `aws_dynamodb_table` | Resource | Tags |
| `aws_elasticache_cluster` | Resource | Auth token (change only), Description, Tags |
| `aws_elasticache_replication_group` | Resource | Auth token (change only), Description, Tags |
| `aws_elasticache_subnet_group` | Connection between resources | Description |
| `aws_rds_cluster` | Resource | Master username, Master password (change only), Tags |
| `aws_rds_cluster_instance` | Resource | Tags |
| `aws_redshift_cluster` | Resource | Master username, Master password (change only), Tags |
| `aws_redshift_subnet_group` | Connection between resources | Description |

## Streaming

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_msk_cluster` | Resource | Tags |
| `aws_msk_configuration` | Resource | Server properties |
| `aws_msk_scram_secret_association` | Connection between resources | — |

## Messaging

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_sns_topic` | Resource | Topic policy, Tags |
| `aws_sns_topic_subscription` | Connection between resources | — |
| `aws_sqs_queue` | Resource | Queue policy, Tags |

## Storage

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_ebs_volume` | Resource | Tags |
| `aws_s3_bucket` | Resource | Bucket policy, Tags |

## Secrets

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_secretsmanager_secret` | Resource | Description, Resource policy, Tags |
| `aws_secretsmanager_secret_version` | Resource | Secret value, Secret binary |

## Encryption

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_kms_key` | Resource | Description, Key policy, Tags |

## Observability

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_cloudwatch_log_group` | Resource | Tags |

## Configuration

| Type | Shown as | Withheld |
| --- | --- | --- |
| `aws_ssm_parameter` | Resource | Value, Value, Description, Tags |
