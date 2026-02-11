package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"netbird-manage/internal/models"
)

// HandleJobsCommand handles peer job-related operations
func (s *Service) HandleJobsCommand(args []string) error {
	cmd := flag.NewFlagSet("job", flag.ContinueOnError)
	cmd.SetOutput(os.Stderr)
	cmd.Usage = PrintJobUsage

	// Query flags
	listFlag := cmd.String("list", "", "List all jobs for a peer (peer ID)")
	inspectFlag := cmd.Bool("inspect", false, "Inspect a specific job")
	outputFlag := cmd.String("output", "table", "Output format: table or json")

	// Create flags
	createFlag := cmd.String("create", "", "Create a bundle collection job for a peer (peer ID)")
	bundleForTimeFlag := cmd.String("bundle-for-time", "", "Time duration for bundle collection")
	logFileCountFlag := cmd.String("log-file-count", "", "Number of log files to collect")
	anonymizeFlag := cmd.Bool("anonymize", false, "Anonymize collected data")

	// Inspect helper flags
	peerIDFlag := cmd.String("peer-id", "", "Peer ID (for --inspect)")
	jobIDFlag := cmd.String("job-id", "", "Job ID (for --inspect)")

	if len(args) == 1 {
		PrintJobUsage()
		return nil
	}

	if err := cmd.Parse(args[1:]); err != nil {
		return nil
	}

	if *listFlag != "" {
		return s.listJobs(*listFlag, *outputFlag)
	}

	if *inspectFlag {
		if *peerIDFlag == "" || *jobIDFlag == "" {
			return fmt.Errorf("--peer-id and --job-id are required for --inspect")
		}
		return s.inspectJob(*peerIDFlag, *jobIDFlag, *outputFlag)
	}

	if *createFlag != "" {
		params := make(map[string]interface{})
		if *bundleForTimeFlag != "" {
			params["bundle_for_time"] = *bundleForTimeFlag
		}
		if *logFileCountFlag != "" {
			count, err := strconv.Atoi(*logFileCountFlag)
			if err != nil {
				return fmt.Errorf("invalid --log-file-count value: %s", *logFileCountFlag)
			}
			params["log_file_count"] = count
		}
		if *anonymizeFlag {
			params["anonymize"] = true
		}
		return s.createJob(*createFlag, params)
	}

	PrintJobUsage()
	return nil
}

// listJobs lists all jobs for a peer
func (s *Service) listJobs(peerID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/jobs", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jobs []models.Job
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(jobs) == 0 {
		fmt.Println("No jobs found for this peer")
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
	fmt.Fprintln(w, "ID\tSTATUS\tTYPE\tCREATED\tCOMPLETED\tTRIGGERED BY")
	fmt.Fprintln(w, "--\t------\t----\t-------\t---------\t------------")

	for _, job := range jobs {
		workloadType := ""
		if job.Workload != nil {
			workloadType = job.Workload.Type
		}
		completedAt := job.CompletedAt
		if completedAt == "" {
			completedAt = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			job.ID,
			job.Status,
			workloadType,
			job.CreatedAt,
			completedAt,
			job.TriggeredBy,
		)
	}
	w.Flush()
	return nil
}

// inspectJob retrieves details of a specific job
func (s *Service) inspectJob(peerID, jobID, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/jobs/"+jobID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var job models.Job
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
	fmt.Println("---------------------------------")
	fmt.Printf("  Status:       %s\n", job.Status)
	fmt.Printf("  Created At:   %s\n", job.CreatedAt)
	if job.CompletedAt != "" {
		fmt.Printf("  Completed At: %s\n", job.CompletedAt)
	}
	fmt.Printf("  Triggered By: %s\n", job.TriggeredBy)
	if job.FailedReason != "" {
		fmt.Printf("  Failed Reason: %s\n", job.FailedReason)
	}

	if job.Workload != nil {
		fmt.Printf("  Workload Type: %s\n", job.Workload.Type)
		if len(job.Workload.Parameters) > 0 {
			fmt.Println("  Parameters:")
			for k, v := range job.Workload.Parameters {
				fmt.Printf("    %s: %v\n", k, v)
			}
		}
		if len(job.Workload.Result) > 0 {
			fmt.Println("  Result:")
			for k, v := range job.Workload.Result {
				fmt.Printf("    %s: %v\n", k, v)
			}
		}
	}

	return nil
}

// createJob creates a new bundle collection job for a peer
func (s *Service) createJob(peerID string, params map[string]interface{}) error {
	req := models.JobCreateRequest{
		Workload: models.JobWorkloadRequest{
			Type:       "bundle",
			Parameters: params,
		},
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/peers/"+peerID+"/jobs", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var job models.Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Job created successfully!\n")
	fmt.Printf("  Job ID:  %s\n", job.ID)
	fmt.Printf("  Status:  %s\n", job.Status)
	if job.Workload != nil {
		fmt.Printf("  Type:    %s\n", job.Workload.Type)
	}
	return nil
}
