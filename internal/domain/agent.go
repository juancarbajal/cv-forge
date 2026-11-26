package domain

// CriticIssue representa una observación o hallazgo semántico detectado por el agente auditor.
type CriticIssue struct {
	Severity   string `json:"severity"` // "WARNING", "BLOCKING"
	Claim      string `json:"claim"`
	Reason     string `json:"reason"`
	Suggestion string `json:"suggestion"`
}

// CriticVerdict representa el veredicto del agente auditor sobre un CV optimizado.
type CriticVerdict struct {
	Approved bool          `json:"approved"`
	Issues   []CriticIssue `json:"issues"`
	Summary  string        `json:"summary"`
}

// GapQuestion representa una pregunta dirigida al candidato para rescatar experiencia no documentada.
type GapQuestion struct {
	RequirementName string `json:"requirement_name"`
	Question        string `json:"question"`
	WhyAsked        string `json:"why_asked"`
}

// GapAnswer representa la respuesta del candidato a una pregunta de gap.
type GapAnswer struct {
	RequirementName string `json:"requirement_name"`
	Question        string `json:"question"`
	Answer          string `json:"answer"`
}

// AgentStepTrace registra cada paso ejecutado dentro del bucle agéntico.
type AgentStepTrace struct {
	Iteration int    `json:"iteration"`
	AgentName string `json:"agent_name"`
	Status    string `json:"status"` // "APPROVED", "REVISE", "FAILED", "COMPLETED"
	Notes     string `json:"notes"`
}
