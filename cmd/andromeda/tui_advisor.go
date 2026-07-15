package main

import (
	"context"
	"strings"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/settingstore"
)

// advisorSystem frames the model as a reviewer giving a second opinion rather than an executor.
const advisorSystem = "You are a senior engineering advisor giving a second opinion. Review the " +
	"conversation and the user's question, then give focused, high-signal guidance: risks, better " +
	"approaches, trade-offs, and what to watch out for. Be concise and direct. Do not use tools; " +
	"respond with advice only."

// advisorAction backs /advisor. A bare invocation (or "model …") is a synchronous config/usage
// reply; a real question runs a one-shot consultation whose answer is shown but NOT added to the
// conversation history — it is a side channel, so the main thread stays clean. The consult uses the
// current provider with the advisor model (settings) or the current model.
func (s *tuiSession) advisorAction(ctx context.Context, args string) string {
	args = strings.TrimSpace(args)
	if sub, rest, _ := strings.Cut(args, " "); sub == "model" {
		return s.setAdvisorModel(strings.TrimSpace(rest))
	}
	if args == "" || args == "model" {
		return s.advisorUsage()
	}
	model := s.advisorModel()
	res, err := app.RunAgent(ctx, app.RunAgentOptions{
		WorkspaceRoot: s.wd,
		Provider:      s.prov,
		Model:         model,
		System:        advisorSystem,
		Goal:          advisorGoal(s.history, args),
	})
	if err != nil {
		return "advisor failed: " + err.Error()
	}
	answer := strings.TrimSpace(res.FinalText)
	if answer == "" {
		return "advisor returned nothing"
	}
	return "advisor · " + model + "\n\n" + answer
}

// advisorModel is the model to consult: the configured advisor model, or the current model.
func (s *tuiSession) advisorModel() string {
	if st, err := settingstore.Load(s.wd); err == nil && st.AdvisorModel != "" {
		return st.AdvisorModel
	}
	return s.cfg.model
}

func (s *tuiSession) setAdvisorModel(name string) string {
	st, _ := settingstore.Load(s.wd)
	st.AdvisorModel = name
	if err := settingstore.Save(s.wd, st); err != nil {
		return "advisor: " + err.Error()
	}
	if name == "" {
		return "advisor model cleared — consultations use the current model (" + s.cfg.model + ")"
	}
	return "advisor model set to " + name + " (must be served by the current provider, " + s.cfg.provider + ")"
}

func (s *tuiSession) advisorUsage() string {
	return "usage: /advisor <question>   ·   consulting: " + s.advisorModel() +
		"\nset a stronger model with  /advisor model <name>  (cleared with /advisor model)"
}

// advisorGoal packages the conversation context and the question for the consultation.
func advisorGoal(history []ports.Message, question string) string {
	conv := renderConversation(history)
	if conv == "" {
		return question
	}
	return "Conversation so far:\n\n" + conv + "\n\n---\n\nMy question: " + question
}
