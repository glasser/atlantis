package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate --package mocks -o mocks/mock_command_requirement_handler.go CommandRequirementHandler
type CommandRequirementHandler interface {
	ValidateProjectDependencies(ctx command.ProjectContext) (string, error)
	ValidatePlanProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateApplyProject(repoDir string, ctx command.ProjectContext) (string, error)
	ValidateImportProject(repoDir string, ctx command.ProjectContext) (string, error)
}

type DefaultCommandRequirementHandler struct {
	WorkingDir WorkingDir
}

func (a *DefaultCommandRequirementHandler) ValidatePlanProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	for _, req := range ctx.PlanRequirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved according to the project's approval rules before running plan.", nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running plan.", nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running plan.", nil
			}
		}
	}
	// Passed all plan requirements configured.
	return "", nil
}

func (a *DefaultCommandRequirementHandler) ValidateApplyProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	for _, req := range ctx.ApplyRequirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved according to the project's approval rules before running apply.", nil
			}
		// this should come before mergeability check since mergeability is a superset of this check.
		case valid.PoliciesPassedCommandReq:
			// We should rely on this function instead of plan status, since plan status after a failed apply will not carry the policy error over.
			if !ctx.PolicyCleared() {
				return "All policies must pass for project before running apply.", nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running apply.", nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running apply.", nil
			}
		}
	}
	// Passed all apply requirements configured.
	return "", nil
}

func (a *DefaultCommandRequirementHandler) ValidateProjectDependencies(ctx command.ProjectContext) (failure string, err error) {
	for _, dependOnProject := range ctx.DependsOn {

		for _, project := range ctx.PullStatus.Projects {

			if project.ProjectName == dependOnProject && project.Status != models.AppliedPlanStatus && project.Status != models.PlannedNoChangesPlanStatus {
				return fmt.Sprintf("Can't apply your project unless you apply its dependencies: [%s]", project.ProjectName), nil
			}
		}
	}

	return "", nil
}

func (a *DefaultCommandRequirementHandler) ValidateImportProject(repoDir string, ctx command.ProjectContext) (failure string, err error) {
	for _, req := range ctx.ImportRequirements {
		switch req {
		case raw.ApprovedRequirement:
			if !ctx.PullReqStatus.ApprovalStatus.IsApproved {
				return "Pull request must be approved according to the project's approval rules before running import.", nil
			}
		case raw.MergeableRequirement:
			if !ctx.PullReqStatus.Mergeable {
				return "Pull request must be mergeable before running import.", nil
			}
		case raw.UnDivergedRequirement:
			if a.WorkingDir.HasDiverged(ctx.Log, repoDir) {
				return "Default branch must be rebased onto pull request before running import.", nil
			}
		}
	}
	// Passed all import requirements configured.
	return "", nil
}
