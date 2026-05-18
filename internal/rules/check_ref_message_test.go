package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
)

func TestCheckRefMessage_Pass(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/ref_message/pass", "../../protos"},
				FilePaths: []string{"models.proto", "refs.proto"},
			},
			RuleIDs: []string{"CLARITY_REF_MESSAGE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: nil,
	}.Run(t)
}

func TestCheckRefMessage_FailMissingFile(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/ref_message/fail_missing_file", "../../protos"},
				FilePaths: []string{"models.proto"},
			},
			RuleIDs: []string{"CLARITY_REF_MESSAGE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: "CLARITY_REF_MESSAGE",
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

func TestCheckRefMessage_FailMissingMessage(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/ref_message/fail_missing_message", "../../protos"},
				FilePaths: []string{"models.proto", "refs.proto"},
			},
			RuleIDs: []string{"CLARITY_REF_MESSAGE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: "CLARITY_REF_MESSAGE",
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

func TestCheckRefMessage_FailWrongRole(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/ref_message/fail_wrong_role", "../../protos"},
				FilePaths: []string{"models.proto", "refs.proto"},
			},
			RuleIDs: []string{"CLARITY_REF_MESSAGE"},
		},
		Spec: &check.Spec{
			Rules: All,
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: "CLARITY_REF_MESSAGE",
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
