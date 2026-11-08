package pricing

import (
	"math"
	"strings"

	"github.com/danushkastanley/tfviz/internal/input"
)

type component struct {
	name, unit, quantity string
	hours                bool
}
type service struct {
	selectors             []string
	base                  component
	usage                 []component
	arnService, arnPrefix string
}

// Fixed, deliberately small billing vocabulary. No catalogue-defined expressions.
var services = map[string]service{
	"aws_instance":              {fields["aws_instance"], component{"instance_hours", "hour", "", true}, nil, "ec2", "instance/"},
	"aws_db_instance":           {fields["aws_db_instance"], component{"instance_hours", "hour", "", true}, nil, "rds", "db:"},
	"aws_ebs_volume":            {[]string{"type"}, component{"storage_gb_month", "GB-month", "size", false}, nil, "ec2", "volume/"},
	"aws_nat_gateway":           {[]string{"connectivity_type"}, component{"gateway_hours", "hour", "", true}, []component{{"processed_gb", "GB", "", false}}, "ec2", "natgateway/"},
	"aws_lb":                    {[]string{"load_balancer_type"}, component{"load_balancer_hours", "hour", "", true}, []component{{"capacity_unit_hours", "capacity-unit-hour", "", false}}, "elasticloadbalancing", "loadbalancer/"},
	"aws_eks_cluster":           {[]string{"version", "upgrade_policy.support_type"}, component{"cluster_hours", "hour", "", true}, nil, "eks", "cluster/"},
	"aws_kms_key":               {[]string{"key_usage", "customer_master_key_spec", "multi_region"}, component{"key_month", "key-month", "", false}, []component{{"requests", "request", "", false}}, "kms", "key/"},
	"aws_secretsmanager_secret": {[]string{}, component{"secret_month", "secret-month", "", false}, []component{{"api_requests", "request", "", false}}, "secretsmanager", "secret:"},
	"aws_elasticache_cluster":   {[]string{"node_type", "engine", "engine_version"}, component{"node_hours", "node-hour", "num_cache_nodes", true}, nil, "elasticache", "cluster:"},
	"aws_redshift_cluster":      {[]string{"node_type", "cluster_type"}, component{"node_hours", "node-hour", "number_of_nodes", true}, nil, "redshift", "cluster:"},
	"aws_vpc_endpoint":          {[]string{"vpc_endpoint_type", "service_name"}, component{"endpoint_hours", "endpoint-hour", "subnet_ids", true}, []component{{"processed_gb", "GB", "", false}}, "ec2", "vpc-endpoint/"},
	"aws_s3_bucket":             {[]string{}, component{}, []component{{"standard_storage_gb_month", "GB-month", "", false}, {"get_requests", "request", "", false}, {"put_requests", "request", "", false}}, "s3", ""},
	"aws_lambda_function":       {[]string{"architecture", "package_type", "memory_size"}, component{}, []component{{"requests", "request", "", false}, {"compute_gb_seconds", "GB-second", "", false}}, "lambda", "function:"},
	"aws_sqs_queue":             {[]string{"fifo_queue"}, component{}, []component{{"requests", "request", "", false}}, "sqs", ""},
	"aws_sns_topic":             {[]string{"fifo_topic"}, component{}, []component{{"publish_requests", "request", "", false}, {"http_deliveries", "delivery", "", false}}, "sns", ""},
	"aws_cloudwatch_log_group":  {[]string{"log_group_class"}, component{}, []component{{"ingested_gb", "GB", "", false}, {"stored_gb_month", "GB-month", "", false}}, "logs", "log-group:"},
}

func selectorKeys(typ string) []string { return services[typ].selectors }
func lookupComponent(typ, name string) (component, bool) {
	s, ok := services[typ]
	if !ok || name == "" {
		return component{}, false
	}
	if s.base.name == name {
		return s.base, true
	}
	for _, c := range s.usage {
		if c.name == name {
			return c, true
		}
	}
	return component{}, false
}
func evidence(v input.Value, key string) input.Value {
	if key == "architecture" {
		a := v.Field("architectures")
		if a.Kind() == input.KindList && len(a.Items()) == 1 {
			return a.Items()[0]
		}
		return a
	}
	return v.Path(strings.Split(key, ".")...)
}
func selector(v input.Value, key string) (string, bool) {
	f := evidence(v, key)
	if key == "multi_az" || key == "multi_region" || key == "fifo_queue" || key == "fifo_topic" {
		b, ok := f.Bool()
		if b {
			return "true", ok
		}
		return "false", ok
	}
	if key == "memory_size" {
		n, ok := f.Number()
		if !ok || n < 128 || n > 10240 || math.Trunc(n) != n {
			return "", false
		}
		return fmtInteger(int64(n)), true
	}
	s, ok := f.String()
	return s, ok && s != ""
}
func validSelectors(typ string, s map[string]string) bool {
	for _, key := range selectorKeys(typ) {
		val := s[key]
		if val == "" || len(val) > 128 || strings.ContainsAny(val, "\x00\r\n") {
			return false
		}
		if (key == "multi_az" || key == "multi_region" || key == "fifo_queue" || key == "fifo_topic") && val != "true" && val != "false" {
			return false
		}
	}
	switch typ {
	case "aws_instance":
		return s["tenancy"] == "default"
	case "aws_nat_gateway":
		return s["connectivity_type"] == "public" || s["connectivity_type"] == "private"
	case "aws_lb":
		return s["load_balancer_type"] == "application" || s["load_balancer_type"] == "network"
	case "aws_eks_cluster":
		return s["upgrade_policy.support_type"] == "STANDARD" || s["upgrade_policy.support_type"] == "EXTENDED"
	case "aws_vpc_endpoint":
		return s["vpc_endpoint_type"] == "Interface"
	case "aws_lambda_function":
		return (s["architecture"] == "x86_64" || s["architecture"] == "arm64") && (s["package_type"] == "Zip" || s["package_type"] == "Image") && validMemory(s["memory_size"])
	}
	return true
}

func baseQuantity(v input.Value, c component) (int64, bool) {
	n := int64(1)
	if c.quantity == "subnet_ids" {
		f := v.Field(c.quantity)
		if f.Kind() != input.KindList || len(f.Items()) == 0 {
			return 0, false
		}
		seen := map[string]bool{}
		for _, item := range f.Items() {
			s, ok := item.String()
			if !ok || s == "" || seen[s] {
				return 0, false
			}
			seen[s] = true
		}
		n = int64(len(seen))
	} else if c.quantity != "" {
		f, ok := v.Field(c.quantity).Number()
		if !ok || f < 1 || f > 1000000 || math.Trunc(f) != f {
			return 0, false
		}
		n = int64(f)
	}
	if c.hours {
		n *= 730
	}
	return n * 1000000, true
}
