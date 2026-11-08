resource "aws_db_subnet_group" "aurora" {
  name       = "platform-aurora"
  subnet_ids = local.private_subnet_ids
}

resource "aws_rds_cluster" "aurora" {
  cluster_identifier     = "platform-aurora"
  engine                 = "aurora-postgresql"
  engine_version         = "17.4"
  master_username        = "platform_admin"
  master_password        = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.aurora.name
  vpc_security_group_ids = [aws_security_group.data.id]
  storage_encrypted      = true
  kms_key_id             = aws_kms_key.platform.arn
  skip_final_snapshot    = true
}

resource "aws_rds_cluster_instance" "aurora" {
  count = local.proposed ? 3 : 2

  identifier         = "platform-aurora-${count.index}"
  cluster_identifier = aws_rds_cluster.aurora.id
  instance_class     = "db.r7g.large"
  engine             = aws_rds_cluster.aurora.engine
}

resource "aws_elasticache_subnet_group" "cache" {
  name       = "platform-cache"
  subnet_ids = local.private_subnet_ids
}

resource "aws_elasticache_replication_group" "cache" {
  replication_group_id       = "platform-cache"
  description                = "Session cache"
  engine                     = "valkey"
  node_type                  = "cache.r7g.large"
  num_cache_clusters         = 2
  subnet_group_name          = aws_elasticache_subnet_group.cache.name
  security_group_ids         = [aws_security_group.data.id]
  transit_encryption_enabled = true
  at_rest_encryption_enabled = true
  auth_token                 = var.cache_auth_token
}

# Proposed-only: once a table is in state, the AWS provider checks it against
# DynamoDB during plan, which fixture generation must never do.
resource "aws_dynamodb_table" "events" {
  count = local.proposed ? 1 : 0

  name         = "platform-events"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"

  attribute {
    name = "pk"
    type = "S"
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.platform.arn
  }
}

resource "aws_s3_bucket" "logs" {
  bucket = "platform-logs-fixture"
}

resource "aws_ebs_volume" "scratch" {
  availability_zone = "eu-west-1a"
  size              = local.proposed ? 40 : 20
  type              = "gp3"
  encrypted         = true
  kms_key_id        = aws_kms_key.platform.arn
}

resource "aws_redshift_subnet_group" "warehouse" {
  name       = "platform-warehouse"
  subnet_ids = local.private_subnet_ids
}

resource "aws_redshift_cluster" "warehouse" {
  cluster_identifier        = "platform-warehouse"
  node_type                 = "ra3.xlplus"
  cluster_type              = "multi-node"
  number_of_nodes           = 2
  database_name             = "platform"
  master_username           = "warehouse_admin"
  master_password           = var.db_password
  cluster_subnet_group_name = aws_redshift_subnet_group.warehouse.name
  vpc_security_group_ids    = [aws_security_group.data.id]
  encrypted                 = true
  kms_key_id                = aws_kms_key.platform.arn
  skip_final_snapshot       = true
}

resource "aws_elasticache_cluster" "legacy" {
  cluster_id         = "platform-legacy"
  engine             = "memcached"
  node_type          = "cache.t4g.small"
  num_cache_nodes    = 1
  subnet_group_name  = aws_elasticache_subnet_group.cache.name
  security_group_ids = [aws_security_group.data.id]
}
