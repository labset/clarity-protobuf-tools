package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
)

func TestCheckEntityFile_Pass(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_file/pass", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FILE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: nil,
	}.Run(t)
}

func TestCheckEntityFile_FailPackageFormat(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_file/fail_package", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FILE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: "LABSET_ENTITY_FILE",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "models.proto",
					StartLine:   7,
					StartColumn: 0,
					EndLine:     11,
					EndColumn:   1,
				},
			},
		},
	}.Run(t)
}

func TestCheckEntityFile_Fail(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/entity_file/fail", "../../protos"},
				FilePaths: []string{"entities.proto"},
			},
			RuleIDs: []string{"LABSET_ENTITY_FILE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  "LABSET_ENTITY_FILE",
				Message: `Message "acme.inventory.v1.Product" with ROLE_ENTITY must be defined in a models.proto file, got "entities.proto".`,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "entities.proto",
					StartLine:   7,
					StartColumn: 0,
					EndLine:     11,
					EndColumn:   1,
				},
			},
		},
	}.Run(t)
}
