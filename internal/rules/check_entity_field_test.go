package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
)

func TestCheckEntityField_Pass(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_field/pass", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FIELD"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: nil,
	}.Run(t)
}

func TestCheckEntityField_FailMissing(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_field/fail_missing", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FIELD"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  "LABSET_ENTITY_FIELD",
				Message: `Message "acme.inventory.v1.Product" with ROLE_ENTITY must have a field named "entity".`,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "models.proto",
					StartLine:   6,
					StartColumn: 0,
					EndLine:     9,
					EndColumn:   1,
				},
			},
		},
	}.Run(t)
}

func TestCheckEntityField_FailNumber(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_field/fail_number", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FIELD"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  "LABSET_ENTITY_FIELD",
				Message: `Message "acme.inventory.v1.Product" field "entity" must be at field number 1, got 2.`,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "models.proto",
					StartLine:   10,
					StartColumn: 2,
					EndLine:     10,
					EndColumn:   35,
				},
			},
		},
	}.Run(t)
}

func TestCheckEntityField_FailType(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_field/fail_type", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FIELD"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  "LABSET_ENTITY_FIELD",
				Message: `Message "acme.inventory.v1.Product" field "entity" must be of type labset.data.v1.Entity, got string.`,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "models.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   20,
				},
			},
		},
	}.Run(t)
}
