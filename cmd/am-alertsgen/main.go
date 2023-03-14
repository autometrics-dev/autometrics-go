package main

import (
	"log"

	_ "github.com/slok/sloth/pkg/prometheus/api/v1"
)

func main() {
	// TODO Replicate these rules from autometrics-rs/autometrics-cli/src/sloth.rs
	//
	// With default values
	//   #[clap(long, default_values = &["90", "95", "99", "99.9"])]
        //   objectives: Vec<Decimal>,
	/*
	 fn generate_success_rate_slo(objective: &Decimal) -> String {
    let objective_fraction = (objective / Decimal::from(100)).normalize();
    let objective_no_decimal = objective.to_string().replace(".", "");

    format!("  - name: success-rate-{objective_no_decimal}
    objective: {objective}
    description: Common SLO based on function success rates
    sli:
      events:
        error_query: sum by (slo_name, objective) (rate(function_calls_count{{objective=\"{objective_fraction}\",result=\"error\"}}[{{{{.window}}}}]))
        total_query: sum by (slo_name, objective) (rate(function_calls_count{{objective=\"{objective_fraction}\"}}[{{{{.window}}}}]))
    alerting:
      name: High Error Rate SLO - {objective}%
      labels:
        category: success-rate
      annotations:
        summary: \"High error rate on SLO: {{{{$labels.slo_name}}}}\"
      page_alert:
        labels:
          severity: page
      ticket_alert:
        labels:
          severity: ticket
")
}

fn generate_latency_slo(objective: &Decimal) -> String {
    let objective_fraction = (objective / Decimal::from(100)).normalize();
    let objective_no_decimal = objective.to_string().replace(".", "");

    format!("  - name: latency-{objective_no_decimal}
    objective: {objective}
    description: Common SLO based on function latency
    sli:
      events:
        error_query: >
          sum by (slo_name, objective) (rate(function_calls_duration_bucket{{objective=\"{objective_fraction}\"}}[{{{{.window}}}}]))
          -
          (sum by (slo_name, objective) (
            label_join(rate(function_calls_duration_bucket{{objective=\"{objective_fraction}\"}}[{{{{.window}}}}]), \"autometrics_check_label_equality\", \"\", \"target_latency\")
            and
            label_join(rate(function_calls_duration_bucket{{objective=\"{objective_fraction}\"}}[{{{{.window}}}}]), \"autometrics_check_label_equality\", \"\", \"le\")
          ))
        total_query: sum by (slo_name, objective) (rate(function_calls_duration_bucket{{objective=\"{objective_fraction}\"}}[{{{{.window}}}}]))
    alerting:
      name: High Latency SLO - {objective}%
      labels:
        category: latency
      annotations:
        summary: \"High latency on SLO: {{{{$labels.slo_name}}}}\"
      page_alert:
        labels:
          severity: page
      ticket_alert:
        labels:
          severity: ticket
")
}
	 */


	// TODO: Once the sloth rules have been made, we should be able to call
	// the "binary" part of the sloth dep to generate the prom rules directly.

	log.Fatalf("unimplemented")
}
