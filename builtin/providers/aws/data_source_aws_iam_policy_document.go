package aws

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

var dataSourceAwsIamPolicyDocumentVarReplacer = strings.NewReplacer("&{", "${")

func dataSourceAwsIamPolicyDocument() *schema.Resource {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		Read: dataSourceAwsIamPolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"statement": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"effect": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Allow",
						},
						"actions":        setOfString,
						"not_actions":    setOfString,
						"resources":      setOfString,
						"not_resources":  setOfString,
						"principals":     dataSourceAwsIamPolicyPrincipalSchema(),
						"not_principals": dataSourceAwsIamPolicyPrincipalSchema(),
						"condition": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"test": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"variable": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"values": &schema.Schema{
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"json": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIamPolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	doc := &IAMPolicyDoc{
		Version: "2012-10-17",
	}

	if policyId, hasPolicyId := d.GetOk("id"); hasPolicyId {
		doc.Id = policyId.(string)
	}

	var cfgStmts = d.Get("statement").(*schema.Set).List()
	stmts := make([]*IAMPolicyStatement, len(cfgStmts))
	doc.Statements = stmts
	for i, stmtI := range cfgStmts {
		cfgStmt := stmtI.(map[string]interface{})
		stmt := &IAMPolicyStatement{
			Effect: cfgStmt["effect"].(string),
		}

		if actions := cfgStmt["actions"].(*schema.Set).List(); len(actions) > 0 {
			stmt.Actions = iamPolicyDecodeConfigStringList(actions)
		}
		if actions := cfgStmt["not_actions"].(*schema.Set).List(); len(actions) > 0 {
			stmt.NotActions = iamPolicyDecodeConfigStringList(actions)
		}

		if resources := cfgStmt["resources"].(*schema.Set).List(); len(resources) > 0 {
			stmt.Resources = dataSourceAwsIamPolicyDocumentReplaceVarsInList(
				iamPolicyDecodeConfigStringList(resources),
			)
		}
		if resources := cfgStmt["not_resources"].(*schema.Set).List(); len(resources) > 0 {
			stmt.NotResources = dataSourceAwsIamPolicyDocumentReplaceVarsInList(
				iamPolicyDecodeConfigStringList(resources),
			)
		}

		if principals := cfgStmt["principals"].(*schema.Set).List(); len(principals) > 0 {
			stmt.Principals = dataSourceAwsIamPolicyDocumentMakePrincipals(principals)
		}

		if principals := cfgStmt["not_principals"].(*schema.Set).List(); len(principals) > 0 {
			stmt.NotPrincipals = dataSourceAwsIamPolicyDocumentMakePrincipals(principals)
		}

		if conditions := cfgStmt["condition"].(*schema.Set).List(); len(conditions) > 0 {
			stmt.Conditions = dataSourceAwsIamPolicyDocumentMakeConditions(conditions)
		}

		stmts[i] = stmt
	}

	jsonDoc, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(hashcode.String(jsonString)))

	return nil
}

func dataSourceAwsIamPolicyDocumentReplaceVarsInList(in []string) []string {
	out := make([]string, len(in))
	for i, item := range in {
		out[i] = dataSourceAwsIamPolicyDocumentVarReplacer.Replace(item)
	}
	return out
}

func dataSourceAwsIamPolicyDocumentMakeConditions(in []interface{}) IAMPolicyStatementConditionSet {
	out := make([]IAMPolicyStatementCondition, len(in))
	for i, itemI := range in {
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementCondition{
			Test:     item["test"].(string),
			Variable: item["variable"].(string),
			Values: dataSourceAwsIamPolicyDocumentReplaceVarsInList(
				iamPolicyDecodeConfigStringList(
					item["values"].(*schema.Set).List(),
				),
			),
		}
	}
	return IAMPolicyStatementConditionSet(out)
}

func dataSourceAwsIamPolicyDocumentMakePrincipals(in []interface{}) IAMPolicyStatementPrincipalSet {
	out := make([]IAMPolicyStatementPrincipal, len(in))
	for i, itemI := range in {
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementPrincipal{
			Type: item["type"].(string),
			Identifiers: dataSourceAwsIamPolicyDocumentReplaceVarsInList(
				iamPolicyDecodeConfigStringList(
					item["identifiers"].(*schema.Set).List(),
				),
			),
		}
	}
	return IAMPolicyStatementPrincipalSet(out)
}

func dataSourceAwsIamPolicyPrincipalSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"identifiers": &schema.Schema{
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}
