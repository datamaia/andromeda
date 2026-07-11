package core

// Permission is a grant to perform a class of side-effecting action (Volume 9, FR-SEC-100).
// Closed enumeration; extending it requires an ADR.
type Permission string

const (
	PermRead                  Permission = "read"
	PermWrite                 Permission = "write"
	PermExecute               Permission = "execute"
	PermNetwork               Permission = "network"
	PermCredentialAccess      Permission = "credential_access"
	PermGitMutation           Permission = "git_mutation"
	PermProcessSpawn          Permission = "process_spawn"
	PermContainerAccess       Permission = "container_access"
	PermExternalServiceAccess Permission = "external_service_access"
	PermClipboard             Permission = "clipboard"
	PermNotifications         Permission = "notifications"
	PermPackageInstallation   Permission = "package_installation"
	PermSystemModification    Permission = "system_modification"
)

// PermissionScope names the extent a grant applies to (Volume 9). Closed enumeration.
type PermissionScope string

const (
	ScopeSession      PermissionScope = "session"
	ScopeWorkspace    PermissionScope = "workspace"
	ScopeCommand      PermissionScope = "command"
	ScopeTool         PermissionScope = "tool"
	ScopeProvider     PermissionScope = "provider"
	ScopeHost         PermissionScope = "host"
	ScopePath         PermissionScope = "path"
	ScopeDomain       PermissionScope = "domain"
	ScopeRepository   PermissionScope = "repository"
	ScopeOrganization PermissionScope = "organization"
)

// PermissionDecisionKind is a user's persisted decision on a permission prompt (Volume 9).
// Closed enumeration.
type PermissionDecisionKind string

const (
	DecisionAllowOnce         PermissionDecisionKind = "allow_once"
	DecisionAllowForSession   PermissionDecisionKind = "allow_for_session"
	DecisionAllowForWorkspace PermissionDecisionKind = "allow_for_workspace"
	DecisionAlwaysAllowPolicy PermissionDecisionKind = "always_allow_policy"
	DecisionDenyOnce          PermissionDecisionKind = "deny_once"
	DecisionAlwaysDeny        PermissionDecisionKind = "always_deny"
	DecisionAskEveryTime      PermissionDecisionKind = "ask_every_time"
)

// DecisionOutcome is the resolved guidance a permission evaluation returns (Volume 9):
// allow proceeds, deny refuses, ask indicates interaction is required.
type DecisionOutcome string

const (
	OutcomeAllow DecisionOutcome = "allow"
	OutcomeDeny  DecisionOutcome = "deny"
	OutcomeAsk   DecisionOutcome = "ask"
)
