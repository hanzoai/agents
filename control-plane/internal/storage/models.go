package storage

import "time"

// Index names are pinned in the tags wherever the engine's default
// (IDX_<table>_<column>) would not land on the name already in the database.
// SQLite matches identifiers case-insensitively, so IDX_/UQE_ and idx_ are the
// same object: the raw CREATE INDEX IF NOT EXISTS statements in local.go stay
// no-ops and no table gains a second copy of an index it already has.

type ExecutionRecordModel struct {
	ID                int64      `xorm:"'id' pk autoincr"`
	ExecutionID       string     `xorm:"'execution_id' notnull unique(IDX_executions_execution_id)"`
	RunID             string     `xorm:"'run_id' notnull index"`
	ParentExecutionID *string    `xorm:"'parent_execution_id' index"`
	AgentNodeID       string     `xorm:"'agent_node_id' notnull index"`
	ReasonerID        string     `xorm:"'reasoner_id' notnull index"`
	NodeID            string     `xorm:"'node_id' notnull index"`
	Status            string     `xorm:"'status' notnull index"`
	InputPayload      []byte     `xorm:"'input_payload'"`
	ResultPayload     []byte     `xorm:"'result_payload'"`
	ErrorMessage      *string    `xorm:"'error_message'"`
	InputURI          *string    `xorm:"'input_uri'"`
	ResultURI         *string    `xorm:"'result_uri'"`
	SessionID         *string    `xorm:"'session_id' index"`
	ActorID           *string    `xorm:"'actor_id' index"`
	StartedAt         time.Time  `xorm:"'started_at' notnull index"`
	CompletedAt       *time.Time `xorm:"'completed_at'"`
	DurationMS        *int64     `xorm:"'duration_ms'"`
	Notes             string     `xorm:"'notes' default('[]')"`
	CreatedAt         time.Time  `xorm:"'created_at'"`
	UpdatedAt         time.Time  `xorm:"'updated_at'"`
}

func (ExecutionRecordModel) TableName() string { return "executions" }

type AgentExecutionModel struct {
	ID           int64     `xorm:"'id' pk autoincr"`
	WorkflowID   string    `xorm:"'workflow_id' notnull index"`
	SessionID    *string   `xorm:"'session_id' index"`
	AgentNodeID  string    `xorm:"'agent_node_id' notnull index"`
	ReasonerID   string    `xorm:"'reasoner_id' notnull index"`
	InputData    []byte    `xorm:"'input_data'"`
	OutputData   []byte    `xorm:"'output_data'"`
	InputSize    int       `xorm:"'input_size'"`
	OutputSize   int       `xorm:"'output_size'"`
	DurationMS   int       `xorm:"'duration_ms' notnull"`
	Status       string    `xorm:"'status' notnull index"`
	ErrorMessage *string   `xorm:"'error_message'"`
	UserID       *string   `xorm:"'user_id'"`
	TeamID       *string   `xorm:"'team_id'"`
	Metadata     []byte    `xorm:"'metadata'"`
	CreatedAt    time.Time `xorm:"'created_at'"`
}

func (AgentExecutionModel) TableName() string { return "agent_executions" }

type AgentNodeModel struct {
	ID                  string     `xorm:"'id' pk"`
	TeamID              string     `xorm:"'team_id' notnull index"`
	BaseURL             string     `xorm:"'base_url' notnull"`
	Version             string     `xorm:"'version' notnull"`
	DeploymentType      string     `xorm:"'deployment_type' default('long_running') index"`
	InvocationURL       *string    `xorm:"'invocation_url'"`
	Reasoners           []byte     `xorm:"'reasoners'"`
	Skills              []byte     `xorm:"'skills'"`
	CommunicationConfig []byte     `xorm:"'communication_config'"`
	HealthStatus        string     `xorm:"'health_status' notnull index"`
	LifecycleStatus     string     `xorm:"'lifecycle_status' default('starting') index"`
	LastHeartbeat       *time.Time `xorm:"'last_heartbeat'"`
	RegisteredAt        time.Time  `xorm:"'registered_at'"`
	Features            []byte     `xorm:"'features'"`
	Metadata            []byte     `xorm:"'metadata'"`
}

func (AgentNodeModel) TableName() string { return "agent_nodes" }

type AgentConfigurationModel struct {
	ID              int64     `xorm:"'id' pk autoincr"`
	AgentID         string    `xorm:"'agent_id' notnull index(IDX_agent_config_agent_package)"`
	PackageID       string    `xorm:"'package_id' notnull index(IDX_agent_config_agent_package)"`
	Configuration   []byte    `xorm:"'configuration' notnull"`
	EncryptedFields []byte    `xorm:"'encrypted_fields'"`
	Status          string    `xorm:"'status' notnull"`
	Version         int       `xorm:"'version' notnull default(1)"`
	CreatedAt       time.Time `xorm:"'created_at'"`
	UpdatedAt       time.Time `xorm:"'updated_at'"`
	CreatedBy       *string   `xorm:"'created_by'"`
	UpdatedBy       *string   `xorm:"'updated_by'"`
}

func (AgentConfigurationModel) TableName() string { return "agent_configurations" }

type AgentPackageModel struct {
	ID                  string    `xorm:"'id' pk"`
	Name                string    `xorm:"'name' notnull"`
	Version             string    `xorm:"'version' notnull"`
	Description         *string   `xorm:"'description'"`
	Author              *string   `xorm:"'author'"`
	Repository          *string   `xorm:"'repository'"`
	InstallPath         string    `xorm:"'install_path' notnull"`
	ConfigurationSchema []byte    `xorm:"'configuration_schema'"`
	Status              string    `xorm:"'status' notnull"`
	ConfigurationStatus string    `xorm:"'configuration_status' notnull"`
	InstalledAt         time.Time `xorm:"'installed_at'"`
	UpdatedAt           time.Time `xorm:"'updated_at'"`
	Metadata            []byte    `xorm:"'metadata'"`
}

func (AgentPackageModel) TableName() string { return "agent_packages" }

type WorkflowExecutionModel struct {
	ID                    int64      `xorm:"'id' pk autoincr"`
	WorkflowID            string     `xorm:"'workflow_id' notnull index index(workflow_status)"`
	ExecutionID           string     `xorm:"'execution_id' notnull unique(IDX_workflow_executions_execution_id)"`
	HanzoAgentsRequestID  string     `xorm:"'hanzo_agents_request_id' notnull index"`
	RunID                 *string    `xorm:"'run_id' index"`
	SessionID             *string    `xorm:"'session_id' index index(session_status) index(session_status_time) index(session_time)"`
	ActorID               *string    `xorm:"'actor_id' index index(actor_status) index(actor_status_time) index(actor_time)"`
	AgentNodeID           string     `xorm:"'agent_node_id' notnull index index(agent_node_status) index(agent_status_time)"`
	ParentWorkflowID      *string    `xorm:"'parent_workflow_id' index"`
	ParentExecutionID     *string    `xorm:"'parent_execution_id' index"`
	RootWorkflowID        *string    `xorm:"'root_workflow_id' index"`
	WorkflowDepth         int        `xorm:"'workflow_depth' default(0)"`
	ReasonerID            string     `xorm:"'reasoner_id' notnull"`
	InputData             []byte     `xorm:"'input_data'"`
	OutputData            []byte     `xorm:"'output_data'"`
	InputSize             int        `xorm:"'input_size'"`
	OutputSize            int        `xorm:"'output_size'"`
	WorkflowName          *string    `xorm:"'workflow_name'"`
	WorkflowTags          string     `xorm:"'workflow_tags'"`
	Status                string     `xorm:"'status' notnull index index(agent_node_status) index(session_status) index(actor_status) index(workflow_status) index(status_time) index(session_status_time) index(actor_status_time) index(agent_status_time)"`
	StartedAt             time.Time  `xorm:"'started_at' notnull index index(status_time) index(session_status_time) index(actor_status_time) index(agent_status_time) index(session_time) index(actor_time)"`
	CompletedAt           *time.Time `xorm:"'completed_at'"`
	DurationMS            int        `xorm:"'duration_ms'"`
	StateVersion          int        `xorm:"'state_version' notnull default(0)"`
	LastEventSequence     int        `xorm:"'last_event_sequence' notnull default(0)"`
	ActiveChildren        int        `xorm:"'active_children' notnull default(0)"`
	PendingChildren       int        `xorm:"'pending_children' notnull default(0)"`
	PendingTerminalStatus *string    `xorm:"'pending_terminal_status'"`
	StatusReason          *string    `xorm:"'status_reason'"`
	LeaseOwner            *string    `xorm:"'lease_owner'"`
	LeaseExpiresAt        *time.Time `xorm:"'lease_expires_at'"`
	ErrorMessage          *string    `xorm:"'error_message'"`
	RetryCount            int        `xorm:"'retry_count' default(0)"`
	Notes                 string     `xorm:"'notes' default('[]')"`
	CreatedAt             time.Time  `xorm:"'created_at'"`
	UpdatedAt             time.Time  `xorm:"'updated_at'"`
}

func (WorkflowExecutionModel) TableName() string { return "workflow_executions" }

type WorkflowExecutionEventModel struct {
	EventID           int64     `xorm:"'event_id' pk autoincr"`
	ExecutionID       string    `xorm:"'execution_id' notnull index(IDX_workflow_exec_events_execution)"`
	WorkflowID        string    `xorm:"'workflow_id' notnull"`
	RunID             *string   `xorm:"'run_id' index(IDX_workflow_exec_events_run)"`
	ParentExecutionID *string   `xorm:"'parent_execution_id'"`
	Sequence          int64     `xorm:"'sequence' notnull index(IDX_workflow_exec_events_execution)"`
	PreviousSequence  int64     `xorm:"'previous_sequence' notnull default(0)"`
	EventType         string    `xorm:"'event_type' notnull"`
	Status            *string   `xorm:"'status'"`
	StatusReason      *string   `xorm:"'status_reason'"`
	Payload           string    `xorm:"'payload' default('{}')"`
	EmittedAt         time.Time `xorm:"'emitted_at' notnull"`
	RecordedAt        time.Time `xorm:"'recorded_at'"`
}

func (WorkflowExecutionEventModel) TableName() string { return "workflow_execution_events" }

type WorkflowRunEventModel struct {
	EventID          int64     `xorm:"'event_id' pk autoincr"`
	RunID            string    `xorm:"'run_id' notnull index(run)"`
	Sequence         int64     `xorm:"'sequence' notnull index(run)"`
	PreviousSequence int64     `xorm:"'previous_sequence' notnull default(0)"`
	EventType        string    `xorm:"'event_type' notnull"`
	Status           *string   `xorm:"'status'"`
	StatusReason     *string   `xorm:"'status_reason'"`
	Payload          string    `xorm:"'payload' default('{}')"`
	EmittedAt        time.Time `xorm:"'emitted_at' notnull"`
	RecordedAt       time.Time `xorm:"'recorded_at'"`
}

func (WorkflowRunEventModel) TableName() string { return "workflow_run_events" }

type WorkflowRunModel struct {
	RunID             string     `xorm:"'run_id' pk"`
	RootWorkflowID    string     `xorm:"'root_workflow_id' notnull index"`
	RootExecutionID   *string    `xorm:"'root_execution_id'"`
	Status            string     `xorm:"'status' notnull default('pending') index"`
	TotalSteps        int        `xorm:"'total_steps' notnull default(0)"`
	CompletedSteps    int        `xorm:"'completed_steps' notnull default(0)"`
	FailedSteps       int        `xorm:"'failed_steps' notnull default(0)"`
	StateVersion      int64      `xorm:"'state_version' notnull default(0)"`
	LastEventSequence int64      `xorm:"'last_event_sequence' notnull default(0)"`
	Metadata          []byte     `xorm:"'metadata' default('{}')"`
	CreatedAt         time.Time  `xorm:"'created_at' index"`
	UpdatedAt         time.Time  `xorm:"'updated_at' index"`
	CompletedAt       *time.Time `xorm:"'completed_at' index"`
}

func (WorkflowRunModel) TableName() string { return "workflow_runs" }

type WorkflowStepModel struct {
	StepID       string     `xorm:"'step_id' pk"`
	RunID        string     `xorm:"'run_id' notnull index index(IDX_workflow_steps_run_execution) index(run_status) index(run_priority)"`
	ParentStepID *string    `xorm:"'parent_step_id' index"`
	ExecutionID  *string    `xorm:"'execution_id' index(IDX_workflow_steps_run_execution)"`
	AgentNodeID  *string    `xorm:"'agent_node_id' index index(agent_not_before)"`
	Target       *string    `xorm:"'target'"`
	Status       string     `xorm:"'status' notnull default('pending') index index(run_status) index(status_not_before) index(agent_not_before)"`
	Attempt      int        `xorm:"'attempt' notnull default(0)"`
	Priority     int        `xorm:"'priority' notnull default(0) index(run_priority)"`
	NotBefore    time.Time  `xorm:"'not_before' notnull index(status_not_before) index(agent_not_before) index(run_priority)"`
	InputURI     *string    `xorm:"'input_uri'"`
	ResultURI    *string    `xorm:"'result_uri'"`
	ErrorMessage *string    `xorm:"'error_message'"`
	Metadata     []byte     `xorm:"'metadata' default('{}')"`
	StartedAt    *time.Time `xorm:"'started_at'"`
	CompletedAt  *time.Time `xorm:"'completed_at'"`
	LeasedAt     *time.Time `xorm:"'leased_at'"`
	LeaseTimeout *time.Time `xorm:"'lease_timeout'"`
	CreatedAt    time.Time  `xorm:"'created_at' index"`
	UpdatedAt    time.Time  `xorm:"'updated_at' index"`
}

func (WorkflowStepModel) TableName() string { return "workflow_steps" }

type WorkflowModel struct {
	WorkflowID           string     `xorm:"'workflow_id' pk"`
	WorkflowName         *string    `xorm:"'workflow_name'"`
	WorkflowTags         string     `xorm:"'workflow_tags'"`
	SessionID            *string    `xorm:"'session_id' index"`
	ActorID              *string    `xorm:"'actor_id' index"`
	ParentWorkflowID     *string    `xorm:"'parent_workflow_id'"`
	ParentExecutionID    *string    `xorm:"'parent_execution_id'"`
	RootWorkflowID       *string    `xorm:"'root_workflow_id'"`
	WorkflowDepth        int        `xorm:"'workflow_depth' default(0)"`
	TotalExecutions      int        `xorm:"'total_executions' default(0)"`
	SuccessfulExecutions int        `xorm:"'successful_executions' default(0)"`
	FailedExecutions     int        `xorm:"'failed_executions' default(0)"`
	TotalDurationMS      int        `xorm:"'total_duration_ms' default(0)"`
	Status               string     `xorm:"'status' notnull"`
	StartedAt            time.Time  `xorm:"'started_at' notnull"`
	CompletedAt          *time.Time `xorm:"'completed_at'"`
	CreatedAt            time.Time  `xorm:"'created_at'"`
	UpdatedAt            time.Time  `xorm:"'updated_at'"`
}

func (WorkflowModel) TableName() string { return "workflows" }

type SessionModel struct {
	SessionID       string    `xorm:"'session_id' pk"`
	ActorID         *string   `xorm:"'actor_id' index"`
	SessionName     *string   `xorm:"'session_name'"`
	ParentSessionID *string   `xorm:"'parent_session_id'"`
	RootSessionID   *string   `xorm:"'root_session_id' index"`
	TotalWorkflows  int       `xorm:"'total_workflows' default(0)"`
	TotalExecutions int       `xorm:"'total_executions' default(0)"`
	TotalDurationMS int       `xorm:"'total_duration_ms' default(0)"`
	StartedAt       time.Time `xorm:"'started_at' notnull"`
	LastActivityAt  time.Time `xorm:"'last_activity_at' notnull"`
	CreatedAt       time.Time `xorm:"'created_at'"`
	UpdatedAt       time.Time `xorm:"'updated_at'"`
}

func (SessionModel) TableName() string { return "sessions" }

type DIDRegistryModel struct {
	HanzoAgentsServerID string    `xorm:"'hanzo_agents_server_id' pk"`
	MasterSeedEncrypted []byte    `xorm:"'master_seed_encrypted' notnull"`
	RootDID             string    `xorm:"'root_did' notnull unique"`
	AgentNodes          string    `xorm:"'agent_nodes' default('{}')"`
	TotalDIDs           int       `xorm:"'total_dids' default(0)"`
	CreatedAt           time.Time `xorm:"'created_at'"`
	LastKeyRotation     time.Time `xorm:"'last_key_rotation'"`
}

func (DIDRegistryModel) TableName() string { return "did_registry" }

type AgentDIDModel struct {
	DID                 string    `xorm:"'did' pk"`
	AgentNodeID         string    `xorm:"'agent_node_id' notnull index"`
	HanzoAgentsServerID string    `xorm:"'hanzo_agents_server_id' notnull index"`
	PublicKeyJWK        string    `xorm:"'public_key_jwk' notnull"`
	DerivationPath      string    `xorm:"'derivation_path' notnull"`
	Reasoners           string    `xorm:"'reasoners' default('{}')"`
	Skills              string    `xorm:"'skills' default('{}')"`
	Status              string    `xorm:"'status' notnull default('active')"`
	RegisteredAt        time.Time `xorm:"'registered_at'"`
	CreatedAt           time.Time `xorm:"'created_at'"`
	UpdatedAt           time.Time `xorm:"'updated_at'"`
}

func (AgentDIDModel) TableName() string { return "agent_dids" }

type ComponentDIDModel struct {
	DID            string    `xorm:"'did' pk"`
	AgentDID       string    `xorm:"'agent_did' notnull index"`
	ComponentType  string    `xorm:"'component_type' notnull index"`
	FunctionName   string    `xorm:"'function_name' notnull"`
	PublicKeyJWK   string    `xorm:"'public_key_jwk' notnull"`
	DerivationPath string    `xorm:"'derivation_path' notnull"`
	Capabilities   string    `xorm:"'capabilities' default('[]')"`
	Tags           string    `xorm:"'tags' default('[]')"`
	ExposureLevel  string    `xorm:"'exposure_level' notnull default('private')"`
	CreatedAt      time.Time `xorm:"'created_at'"`
	UpdatedAt      time.Time `xorm:"'updated_at'"`
}

func (ComponentDIDModel) TableName() string { return "component_dids" }

type ExecutionVCModel struct {
	VCID              string    `xorm:"'vc_id' pk"`
	ExecutionID       string    `xorm:"'execution_id' notnull index index(IDX_execution_vcs_execution_unique)"`
	WorkflowID        string    `xorm:"'workflow_id' notnull index"`
	SessionID         string    `xorm:"'session_id' notnull index"`
	IssuerDID         string    `xorm:"'issuer_did' notnull index index(IDX_execution_vcs_execution_unique)"`
	TargetDID         *string   `xorm:"'target_did' index index(IDX_execution_vcs_execution_unique)"`
	CallerDID         string    `xorm:"'caller_did' notnull index"`
	VCDocument        string    `xorm:"'vc_document' notnull"`
	Signature         string    `xorm:"'signature' notnull"`
	StorageURI        string    `xorm:"'storage_uri' default('')"`
	DocumentSizeBytes int64     `xorm:"'document_size_bytes' default(0)"`
	InputHash         string    `xorm:"'input_hash' notnull"`
	OutputHash        string    `xorm:"'output_hash' notnull"`
	Status            string    `xorm:"'status' notnull default('pending') index"`
	ParentVCID        *string   `xorm:"'parent_vc_id' index"`
	ChildVCIDs        string    `xorm:"'child_vc_ids' default('[]')"`
	CreatedAt         time.Time `xorm:"'created_at' index"`
	UpdatedAt         time.Time `xorm:"'updated_at'"`
}

func (ExecutionVCModel) TableName() string { return "execution_vcs" }

type WorkflowVCModel struct {
	WorkflowVCID      string     `xorm:"'workflow_vc_id' pk"`
	WorkflowID        string     `xorm:"'workflow_id' notnull index"`
	SessionID         string     `xorm:"'session_id' notnull index"`
	ComponentVCIDs    string     `xorm:"'component_vc_ids' default('[]')"`
	Status            string     `xorm:"'status' notnull default('pending') index"`
	StartTime         time.Time  `xorm:"'start_time' index"`
	EndTime           *time.Time `xorm:"'end_time' index"`
	TotalSteps        int        `xorm:"'total_steps' default(0)"`
	CompletedSteps    int        `xorm:"'completed_steps' default(0)"`
	StorageURI        string     `xorm:"'storage_uri' default('')"`
	DocumentSizeBytes int64      `xorm:"'document_size_bytes' default(0)"`
	CreatedAt         time.Time  `xorm:"'created_at' index"`
	UpdatedAt         time.Time  `xorm:"'updated_at'"`
}

func (WorkflowVCModel) TableName() string { return "workflow_vcs" }

type SchemaMigrationModel struct {
	Version     string    `xorm:"'version' pk"`
	AppliedAt   time.Time `xorm:"'applied_at'"`
	Description string    `xorm:"'description'"`
}

func (SchemaMigrationModel) TableName() string { return "schema_migrations" }

type ExecutionWebhookEventModel struct {
	ID           int64     `xorm:"'id' pk autoincr"`
	ExecutionID  string    `xorm:"'execution_id' notnull index"`
	EventType    string    `xorm:"'event_type' notnull"`
	Status       string    `xorm:"'status' notnull"`
	HTTPStatus   *int      `xorm:"'http_status'"`
	Payload      *string   `xorm:"'payload'"`
	ResponseBody *string   `xorm:"'response_body'"`
	ErrorMessage *string   `xorm:"'error_message'"`
	CreatedAt    time.Time `xorm:"'created_at'"`
}

func (ExecutionWebhookEventModel) TableName() string { return "execution_webhook_events" }

type ExecutionWebhookModel struct {
	ExecutionID   string     `xorm:"'execution_id' pk"`
	URL           string     `xorm:"'url' notnull"`
	Secret        *string    `xorm:"'secret'"`
	Headers       string     `xorm:"'headers' default('{}')"`
	Status        string     `xorm:"'status' notnull default('pending')"`
	AttemptCount  int        `xorm:"'attempt_count' notnull default(0)"`
	NextAttemptAt *time.Time `xorm:"'next_attempt_at'"`
	LastAttemptAt *time.Time `xorm:"'last_attempt_at'"`
	LastError     *string    `xorm:"'last_error'"`
	CreatedAt     time.Time  `xorm:"'created_at'"`
	UpdatedAt     time.Time  `xorm:"'updated_at'"`
}

func (ExecutionWebhookModel) TableName() string { return "execution_webhooks" }

// ObservabilityWebhookModel represents the global observability webhook configuration.
// This is a singleton table with only one row (id='global').
type ObservabilityWebhookModel struct {
	ID        string    `xorm:"'id' pk default('global')"`
	URL       string    `xorm:"'url' notnull"`
	Secret    *string   `xorm:"'secret'"`
	Headers   string    `xorm:"'headers' default('{}')"`
	Enabled   bool      `xorm:"'enabled' notnull default(true)"`
	CreatedAt time.Time `xorm:"'created_at'"`
	UpdatedAt time.Time `xorm:"'updated_at'"`
}

func (ObservabilityWebhookModel) TableName() string { return "observability_webhooks" }

// ObservabilityDeadLetterQueueModel represents failed observability events for retry.
type ObservabilityDeadLetterQueueModel struct {
	ID             int64     `xorm:"'id' pk autoincr"`
	EventType      string    `xorm:"'event_type' notnull"`
	EventSource    string    `xorm:"'event_source' notnull"`
	EventTimestamp time.Time `xorm:"'event_timestamp' notnull"`
	Payload        string    `xorm:"'payload' notnull"`
	ErrorMessage   string    `xorm:"'error_message' notnull"`
	RetryCount     int       `xorm:"'retry_count' notnull default(0)"`
	CreatedAt      time.Time `xorm:"'created_at'"`
}

func (ObservabilityDeadLetterQueueModel) TableName() string { return "observability_dead_letter_queue" }
