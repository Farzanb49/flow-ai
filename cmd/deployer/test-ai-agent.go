package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestAIAgent tests the AI agent functionality
func TestAIAgent(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	// Test error detection
	testCases := []struct {
		name     string
		logLine  string
		expected string
	}{
		{
			name:     "Buildpack Failure",
			logLine:  "ERROR: failed to build: failed to fetch base layers",
			expected: "buildpack_failure",
		},
		{
			name:     "ECR Auth Failed",
			logLine:  "authentication failed: unauthorized",
			expected: "ecr_auth_failed",
		},
		{
			name:     "Repository Not Found",
			logLine:  "repository not found: no such repository",
			expected: "repository_not_found",
		},
		{
			name:     "Revision Missing",
			logLine:  "revision missing: revision not found",
			expected: "revision_missing",
		},
		{
			name:     "Image Pull Backoff",
			logLine:  "imagepullbackoff: failed to pull image",
			expected: "image_pull_backoff",
		},
		{
			name:     "DNS Resolution Failed",
			logLine:  "could not resolve host: dns failed",
			expected: "dns_resolution_failed",
		},
		{
			name:     "Permission Denied",
			logLine:  "permission denied: access denied",
			expected: "permission_denied",
		},
		{
			name:     "LoadBalancer Pending",
			logLine:  "loadbalancer pending: external pending",
			expected: "loadbalancer_pending",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			error := agent.detectError(tc.logLine)
			if error == nil {
				t.Errorf("Expected to detect error for: %s", tc.logLine)
				return
			}
			if error.Pattern.Name != tc.expected {
				t.Errorf("Expected error pattern %s, got %s", tc.expected, error.Pattern.Name)
			}
		})
	}
}

// TestFixStrategies tests the fix strategy selection
func TestFixStrategies(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	// Test buildpack failure fix
	error := &DeploymentError{
		Pattern: &ErrorPattern{
			Name: "buildpack_failure",
		},
		Suggestions: agent.fixStrategies["buildpack_failure"],
	}

	bestFix := agent.selectBestFix(error)
	if bestFix == nil {
		t.Error("Expected to find a fix strategy for buildpack failure")
		return
	}

	if bestFix.Name != "docker_fallback" {
		t.Errorf("Expected docker_fallback fix, got %s", bestFix.Name)
	}

	if bestFix.Confidence < 0.8 {
		t.Errorf("Expected high confidence for docker_fallback, got %.2f", bestFix.Confidence)
	}
}

// TestLogMonitoring tests the log monitoring functionality
func TestLogMonitoring(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	// Create a test log reader
	logContent := `Starting build...
Building with pack...
ERROR: failed to build: failed to fetch base layers
Falling back to Docker build...
Docker build successful!
Deploying to cluster...
Success!`

	reader := strings.NewReader(logContent)
	
	// Monitor logs
	agent.MonitorLogs(reader)
	
	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)
	
	// Check if error was detected
	logs := agent.ExportLogs()
	if len(logs) == 0 {
		t.Error("Expected logs to be captured")
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	// Test with nil error
	bestFix := agent.selectBestFix(nil)
	if bestFix != nil {
		t.Error("Expected nil fix for nil error")
	}

	// Test with empty suggestions
	error := &DeploymentError{
		Pattern: &ErrorPattern{
			Name: "unknown_error",
		},
		Suggestions: []FixStrategy{},
	}

	bestFix = agent.selectBestFix(error)
	if bestFix != nil {
		t.Error("Expected nil fix for empty suggestions")
	}
}

// TestAIAgentIntegration tests the AI agent integration
func TestAIAgentIntegration(t *testing.T) {
	// This would test the full integration with the deploy command
	// For now, we'll test the basic functionality
	
	agent := NewAIAgent()
	defer agent.Stop()

	// Test status
	status := agent.GetStatus()
	if status.Status != "monitoring" {
		t.Errorf("Expected status 'monitoring', got %s", status.Status)
	}

	// Test log export
	logs := agent.ExportLogs()
	if logs == nil {
		t.Error("Expected logs to be non-nil")
	}
}

// BenchmarkErrorDetection benchmarks error detection performance
func BenchmarkErrorDetection(b *testing.B) {
	agent := NewAIAgent()
	defer agent.Stop()

	logLine := "ERROR: failed to build: failed to fetch base layers"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.detectError(logLine)
	}
}

// ExampleAIAgent demonstrates how to use the AI agent
func ExampleAIAgent() {
	// Create a new AI agent
	agent := NewAIAgent()
	defer agent.Stop()

	// Simulate monitoring logs
	logContent := `Building application...
ERROR: failed to build: buildpack failure
Switching to Docker build...
Build successful!`

	reader := strings.NewReader(logContent)
	agent.MonitorLogs(reader)

	// Get status
	status := agent.GetStatus()
	fmt.Printf("Agent status: %s\n", status.Status)
	
	// Export logs
	logs := agent.ExportLogs()
	fmt.Printf("Captured %d log lines\n", len(logs))
}

// TestRealWorldScenarios tests real-world error scenarios
func TestRealWorldScenarios(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	scenarios := []struct {
		name    string
		logs    []string
		expectError bool
	}{
		{
			name: "Docker Build Success",
			logs: []string{
				"Building application...",
				"Using Docker build...",
				"Build successful!",
			},
			expectError: false,
		},
		{
			name: "Buildpack Failure with Recovery",
			logs: []string{
				"Building with pack...",
				"ERROR: failed to build: failed to fetch base layers",
				"Falling back to Docker build...",
				"Docker build successful!",
			},
			expectError: true,
		},
		{
			name: "ECR Push Failure",
			logs: []string{
				"Pushing to ECR...",
				"ERROR: push failed: authentication failed",
				"Refreshing ECR auth...",
				"Push successful!",
			},
			expectError: true,
		},
		{
			name: "Knative Deploy Failure",
			logs: []string{
				"Deploying to Knative...",
				"ERROR: revision missing",
				"Restarting service...",
				"Deploy successful!",
			},
			expectError: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Reset agent for each scenario
			agent = NewAIAgent()
			defer agent.Stop()

			// Simulate logs
			logContent := strings.Join(scenario.logs, "\n")
			reader := strings.NewReader(logContent)
			agent.MonitorLogs(reader)

			// Wait for processing
			time.Sleep(100 * time.Millisecond)

			// Check if error was detected
			logs := agent.ExportLogs()
			errorDetected := false
			for _, log := range logs {
				if agent.detectError(log) != nil {
					errorDetected = true
					break
				}
			}

			if scenario.expectError && !errorDetected {
				t.Errorf("Expected error to be detected in scenario: %s", scenario.name)
			}
			if !scenario.expectError && errorDetected {
				t.Errorf("Expected no error in scenario: %s", scenario.name)
			}
		})
	}
}

// TestFixStrategyConfidence tests fix strategy confidence levels
func TestFixStrategyConfidence(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	// Test that all fix strategies have reasonable confidence levels
	for errorType, strategies := range agent.fixStrategies {
		for i, strategy := range strategies {
			if strategy.Confidence < 0.0 || strategy.Confidence > 1.0 {
				t.Errorf("Invalid confidence level for %s strategy %d: %.2f", errorType, i, strategy.Confidence)
			}
			if strategy.Confidence < 0.5 {
				t.Logf("Warning: Low confidence for %s strategy %s: %.2f", errorType, strategy.Name, strategy.Confidence)
			}
		}
	}
}

// TestErrorPatternMatching tests error pattern matching accuracy
func TestErrorPatternMatching(t *testing.T) {
	agent := NewAIAgent()
	defer agent.Stop()

	// Test positive cases
	positiveCases := []struct {
		pattern string
		text    string
	}{
		{"buildpack_failure", "ERROR: failed to build: failed to fetch base layers"},
		{"ecr_auth_failed", "authentication failed: unauthorized"},
		{"repository_not_found", "repository not found: no such repository"},
		{"revision_missing", "revision missing: revision not found"},
		{"image_pull_backoff", "imagepullbackoff: failed to pull image"},
		{"dns_resolution_failed", "could not resolve host: dns failed"},
		{"permission_denied", "permission denied: access denied"},
		{"loadbalancer_pending", "loadbalancer pending: external pending"},
	}

	for _, tc := range positiveCases {
		t.Run(tc.pattern, func(t *testing.T) {
			error := agent.detectError(tc.text)
			if error == nil {
				t.Errorf("Expected to detect %s for text: %s", tc.pattern, tc.text)
			} else if error.Pattern.Name != tc.pattern {
				t.Errorf("Expected pattern %s, got %s", tc.pattern, error.Pattern.Name)
			}
		})
	}

	// Test negative cases (should not match)
	negativeCases := []string{
		"Building application successfully",
		"Deploy completed without errors",
		"All checks passed",
		"Service is running normally",
	}

	for _, text := range negativeCases {
		t.Run("negative_"+text, func(t *testing.T) {
			error := agent.detectError(text)
			if error != nil {
				t.Errorf("Expected no error detection for: %s", text)
			}
		})
	}
}
