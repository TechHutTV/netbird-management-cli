// jobs.go - Asynchronous peer job management (e.g., debug bundle collection)
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"netbird-manage/internal/models"
)

// HandleJobsCommand routes peer job commands
func (s *Service) HandleJobsCommand(args []string) error {
	jobCmd := flag.NewFlagSet("job", flag.ContinueOnError)
	jobCmd.SetOutput(os.Stderr)
	jobCmd.Usage = PrintJobUsage

	// Query flags
	listFlag := jobCmd.Bool("list", false, "List jobs for a peer (requires --peer)")
	inspectFlag := jobCmd.String("inspect", "", "Inspect a job by its ID (requires --peer)")

	// Create flags
	createFlag := jobCmd.Bool("create", false, "Create a job for a peer (requires --peer)")
	typeFlag := jobCmd.String("type", "bundle", "Job workload type (currently only: bundle)")
	bundleForTimeFlag := jobCmd.Int("bundle-for-time", 0, "Collect the bundle over this many seconds")
	logFileCountFlag := jobCmd.Int("log-file-count", 0, "Number of rotated log files to include")
	anonymizeFlag := jobCmd.Bool("anonymize", false, "Anonymize the collected bundle")

	// Common parameters
	peerFlag := jobCmd.String("peer", "", "Peer ID (required for all operations)")

	// Output format
	outputFlag := jobCmd.String("output", "table", "Output format: table or json")

	if len(args) == 1 {
		PrintJobUsage()
		return nil
	}

	if err := jobCmd.Parse(args[1:]); err != nil {
		return err
	}

	if *peerFlag == "" && (*listFlag || *inspectFlag != "" || *createFlag) {
		return fmt.Errorf("--peer is required")
	}

	if *listFlag {
		return s.listPeerJobs(*peerFlag, *outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectPeerJob(*peerFlag, *inspectFlag, *outputFlag)
	}

	if *createFlag {
		if *typeFlag != "bundle" {
			return fmt.Errorf("unsupported job type %q (currently only: bundle)", *typeFlag)
		}
		req := models.PeerJobCreateRequest{
			Workload: models.JobWorkload{
				Type: *typeFlag,
				Parameters: &models.BundleJobParameters{
					BundleFor:     *bundleForTimeFlag > 0,
					BundleForTime: *bundleForTimeFlag,
					LogFileCount:  *logFileCountFlag,
					Anonymize:     *anonymizeFlag,
				},
			},
		}
		return s.createPeerJob(*peerFlag, req)
	}

	jobCmd.Usage()
	return nil
}

// listPeerJobs lists all jobs for a peer
func (s *Service) listPeerJobs(peerID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/jobs", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jobs []models.PeerJob
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(jobs) == 0 {
		if outputFormat == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("No jobs found for this peer")
		}
		return nil
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(jobs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tCREATED AT\tCOMPLETED AT\tTRIGGERED BY")
	fmt.Fprintln(w, "--\t----\t------\t----------\t------------\t------------")

	for _, job := range jobs {
		completedAt := job.CompletedAt
		if completedAt == "" {
			completedAt = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			job.ID,
			job.Workload.Type,
			job.Status,
			job.CreatedAt,
			completedAt,
			job.TriggeredBy,
		)
	}
	w.Flush()
	return nil
}

// inspectPeerJob shows detailed information about a job
func (s *Service) inspectPeerJob(peerID, jobID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/jobs/"+jobID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var job models.PeerJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Job: %s\n", job.ID)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Type:         %s\n", job.Workload.Type)
	fmt.Printf("  Status:       %s\n", job.Status)
	fmt.Printf("  Created At:   %s\n", job.CreatedAt)
	if job.CompletedAt != "" {
		fmt.Printf("  Completed At: %s\n", job.CompletedAt)
	}
	if job.TriggeredBy != "" {
		fmt.Printf("  Triggered By: %s\n", job.TriggeredBy)
	}
	if job.FailedReason != "" {
		fmt.Printf("  Failed:       %s\n", job.FailedReason)
	}
	if job.Workload.Parameters != nil {
		params := job.Workload.Parameters
		fmt.Println("  Parameters:")
		fmt.Printf("    Bundle For:      %t (%ds)\n", params.BundleFor, params.BundleForTime)
		fmt.Printf("    Log File Count:  %d\n", params.LogFileCount)
		fmt.Printf("    Anonymize:       %t\n", params.Anonymize)
	}
	if job.Workload.Result != nil && job.Workload.Result.UploadKey != "" {
		fmt.Printf("  Upload Key:   %s\n", job.Workload.Result.UploadKey)
	}
	return nil
}

// createPeerJob creates a new job for a peer
func (s *Service) createPeerJob(peerID string, req models.PeerJobCreateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/peers/"+peerID+"/jobs", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var job models.PeerJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Job created successfully\n")
	fmt.Printf("  ID:     %s\n", job.ID)
	fmt.Printf("  Type:   %s\n", job.Workload.Type)
	fmt.Printf("  Status: %s\n", job.Status)
	return nil
}
