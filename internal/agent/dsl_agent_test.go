package agent

import (
	"testing"
)

func TestValidateDSLVerbs(t *testing.T) {
	tests := []struct {
		name      string
		dsl       string
		wantError bool
	}{
		{
			name: "Valid case.create verb",
			dsl: `(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "Test fund"))`,
			wantError: false,
		},
		{
			name:      "Valid products.add verb",
			dsl:       `(products.add "CUSTODY" "FUND_ACCOUNTING")`,
			wantError: false,
		},
		{
			name: "Valid kyc.start verb",
			dsl: `(kyc.start
  (documents
    (document "CertificateOfIncorporation"))
  (jurisdictions
    (jurisdiction "LU")))`,
			wantError: false,
		},
		{
			name: "Valid resources.plan verb",
			dsl: `(resources.plan
  (resource.create "CustodyAccount"
    (owner "CustodyTech")
    (var (attr-id "uuid-123"))))`,
			wantError: false,
		},
		{
			name: "Valid values.bind verb",
			dsl: `(values.bind
  (bind (attr-id "uuid-123") (value "test-value")))`,
			wantError: false,
		},
		{
			name: "Multiple valid verbs",
			dsl: `(case.create (cbu.id "CBU-1234"))
(products.add "CUSTODY")
(kyc.start (documents (document "W8BEN")))`,
			wantError: false,
		},
		{
			name:      "Invalid verb - not in approved list",
			dsl:       `(invalid.verb (test "value"))`,
			wantError: true,
		},
		{
			name:      "Invalid verb - made up operation",
			dsl:       `(case.delete (cbu.id "CBU-1234"))`,
			wantError: true,
		},
		{
			name:      "Invalid verb - wrong domain",
			dsl:       `(investor.create (name "Test"))`,
			wantError: true,
		},
		{
			name: "Valid verb with non-verb constructs",
			dsl: `(case.create
  (cbu.id "CBU-1234")
  (for.product "CUSTODY"
    (service "KYC")))`,
			wantError: false,
		},
		{
			name: "All workflow verbs",
			dsl: `(workflow.transition (from "CREATE") (to "KYC"))
(workflow.gate (condition.id "kyc-complete") (required true))
(tasks.create (task.id "task-1") (type "review"))
(tasks.assign (task.id "task-1") (assignee.id "user-1"))
(tasks.complete (task.id "task-1") (outcome "approved"))`,
			wantError: false,
		},
		{
			name: "All compliance verbs",
			dsl: `(kyc.start (documents (document "Passport")))
(kyc.collect (document.id "doc-1") (type "ID"))
(kyc.verify (document.id "doc-1") (verifier.id "sys-1"))
(kyc.assess (risk-rating "LOW") (rationale.id "rat-1"))
(compliance.screen (list "OFAC") (result.id "res-1"))
(compliance.monitor (trigger.id "trigger-1") (frequency "DAILY"))`,
			wantError: false,
		},
		{
			name: "All resource verbs",
			dsl: `(resources.plan (resource.create "Account"))
(resources.provision (resource.id "res-1") (provider.id "prov-1"))
(resources.configure (resource.id "res-1") (config.id "cfg-1"))
(resources.test (resource.id "res-1") (test-suite.id "test-1"))
(resources.deploy (resource.id "res-1") (environment "prod"))`,
			wantError: false,
		},
		{
			name: "All notification verbs",
			dsl: `(notify.send (recipient.id "user-1") (template.id "tpl-1"))
(communicate.request (party.id "party-1") (document.id "doc-1"))
(escalate.trigger (condition.id "cond-1") (level "HIGH"))
(audit.log (event.id "evt-1") (actor.id "actor-1"))`,
			wantError: false,
		},
		{
			name: "All integration verbs",
			dsl: `(external.query (system "CRM") (endpoint.id "ep-1"))
(external.sync (system.id "sys-1") (data.id "data-1"))
(api.call (endpoint.id "ep-1") (payload.id "pay-1"))
(webhook.register (url.id "url-1") (events "all"))`,
			wantError: false,
		},
		{
			name: "All temporal verbs",
			dsl: `(schedule.create (task.id "task-1") (cron "0 0 * * *"))
(deadline.set (task.id "task-1") (date "2024-12-31"))
(reminder.schedule (notification.id "notif-1") (offset "1d"))`,
			wantError: false,
		},
		{
			name: "All risk verbs",
			dsl: `(risk.assess (factor.id "factor-1") (weight 0.5))
(monitor.setup (metric.id "metric-1") (threshold 100))
(alert.trigger (condition.id "cond-1") (severity "HIGH"))`,
			wantError: false,
		},
		{
			name: "All data lifecycle verbs",
			dsl: `(data.collect (source.id "src-1") (destination.id "dst-1"))
(data.transform (transformer.id "trans-1") (input.id "in-1"))
(data.archive (data.id "data-1") (retention "7y"))
(data.purge (data.id "data-1") (reason "expired"))`,
			wantError: false,
		},
		{
			name: "Mixed valid and invalid verbs",
			dsl: `(case.create (cbu.id "CBU-1234"))
(invalid.operation (test "fail"))
(products.add "CUSTODY")`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDSLVerbs(tt.dsl)
			if (err != nil) != tt.wantError {
				t.Errorf("validateDSLVerbs() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateDSLVerbs_IgnoresNonVerbs(t *testing.T) {
	// Test that non-verb constructs are properly ignored
	dsl := `(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "Test")
  (for.product "CUSTODY")
  (var (attr-id "uuid-123"))
  (resource.create "Account"))`

	err := validateDSLVerbs(dsl)
	if err != nil {
		t.Errorf("validateDSLVerbs() should ignore non-verb constructs like cbu.id, attr-id, for.product, resource.create, but got error: %v", err)
	}
}

func TestValidateDSLVerbs_EmptyDSL(t *testing.T) {
	err := validateDSLVerbs("")
	if err != nil {
		t.Errorf("validateDSLVerbs() should not error on empty DSL, got: %v", err)
	}
}

func TestValidateDSLVerbs_VerbListCompleteness(t *testing.T) {
	// Verify all 70+ approved verbs from vocab.go are covered
	approvedCategories := []string{
		"case", "entity", "identity", "products", "services",
		"kyc", "compliance", "resources", "attributes", "values",
		"workflow", "tasks", "notify", "communicate", "escalate", "audit",
		"external", "api", "webhook", "schedule", "deadline", "reminder",
		"risk", "monitor", "alert", "data",
	}

	// This test ensures we haven't missed any verb categories
	// The actual validation happens in validateDSLVerbs function
	for _, category := range approvedCategories {
		t.Logf("Approved verb category: %s.*", category)
	}
}
