locals {
  subnets = {
    a = { az = "eu-west-1a", cidr = "10.50.10.0/24" }
    b = { az = "eu-west-1b", cidr = "10.50.11.0/24" }
  }
}

resource "aws_vpc" "platform" {
  cidr_block = "10.50.0.0/16"
  tags       = { Name = "platform" }
}

resource "aws_subnet" "private" {
  for_each = local.subnets

  vpc_id            = aws_vpc.platform.id
  availability_zone = each.value.az
  cidr_block        = each.value.cidr
  tags              = { Name = "platform-private-${each.key}" }
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.platform.id
}

resource "aws_route_table_association" "private" {
  for_each = local.subnets

  subnet_id      = aws_subnet.private[each.key].id
  route_table_id = aws_route_table.private.id
}

resource "aws_security_group" "workloads" {
  name        = "platform-workloads"
  description = "Platform workloads"
  vpc_id      = aws_vpc.platform.id
}

resource "aws_security_group" "data" {
  name        = "platform-data"
  description = "Platform data stores"
  vpc_id      = aws_vpc.platform.id
}

resource "aws_kms_key" "platform" {
  description = "Platform data encryption"
}

resource "aws_cloudwatch_log_group" "apps" {
  name              = "/platform/apps"
  retention_in_days = 30
  kms_key_id        = aws_kms_key.platform.arn
}

# Unsupported by design: IAM appears only as a limited generic card.
resource "aws_iam_role" "workloads" {
  name = "platform-workloads"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [{ Effect = "Allow", Action = "sts:AssumeRole", Principal = { Service = ["ecs-tasks.amazonaws.com", "lambda.amazonaws.com", "eks.amazonaws.com", "ec2.amazonaws.com"] } }]
  })
}

locals {
  private_subnet_ids = [aws_subnet.private["a"].id, aws_subnet.private["b"].id]
}

resource "aws_vpc_endpoint" "sqs" {
  vpc_id              = aws_vpc.platform.id
  service_name        = "com.amazonaws.eu-west-1.sqs"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = local.private_subnet_ids
  security_group_ids  = [aws_security_group.workloads.id]
  private_dns_enabled = false
}

resource "aws_internet_gateway" "platform" {
  vpc_id = aws_vpc.platform.id
}

resource "aws_route_table" "edge" {
  vpc_id = aws_vpc.platform.id
}

resource "aws_route" "edge_default" {
  route_table_id         = aws_route_table.edge.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.platform.id
}

resource "aws_vpc_security_group_egress_rule" "workloads_https" {
  security_group_id = aws_security_group.workloads.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
}

resource "aws_security_group_rule" "data_from_workloads" {
  type                     = "ingress"
  security_group_id        = aws_security_group.data.id
  source_security_group_id = aws_security_group.workloads.id
  protocol                 = "tcp"
  from_port                = 5432
  to_port                  = 5432
}
