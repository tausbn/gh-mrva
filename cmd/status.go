/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
  "log"

	"github.com/spf13/cobra"
  "github.com/GitHubSecurityLab/gh-mrva/utils"
  "github.com/GitHubSecurityLab/gh-mrva/models"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Checks the status of a given session.",
	Long: `Checks the status of a given session.`,
	Run: func(cmd *cobra.Command, args []string) {
    sessionStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVarP(&sessionNameFlag, "session", "s", "", "Session name be deleted")
	statusCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format (default: false)")
	statusCmd.MarkFlagRequired("session")
}

func sessionStatus() {

	controller, runs, _, err := utils.LoadSession(sessionNameFlag)
	if err != nil {
		log.Fatal(err)
	}
	if len(runs) == 0 {
		log.Fatal("No runs found for run name", sessionNameFlag)
	}


	var results models.Results

	for _, run := range runs {
		if err != nil {
			log.Fatal(err)
		}
		runDetails, err := utils.GetRunDetails(controller, run.Id)
		if err != nil {
			log.Fatal(err)
		}

		status := runDetails["status"].(string)
		var failure_reason string
		if status == "failed" {
			failure_reason = runDetails["failure_reason"].(string)
		} else {
			failure_reason = ""
		}

		results.Runs = append(results.Runs, models.RunStatus{
			Id:            run.Id,
			Query:         run.Query,
			Status:        status,
			FailureReason: failure_reason,
		})

		for _, repo := range runDetails["scanned_repositories"].([]interface{}) {
			if repo.(map[string]interface{})["analysis_status"].(string) == "succeeded" {
				results.TotalSuccessfulScans += 1
				if repo.(map[string]interface{})["result_count"].(float64) > 0 {
					results.TotalRepositoriesWithFindings += 1
					results.TotalFindingsCount += int(repo.(map[string]interface{})["result_count"].(float64))
					repoInfo := repo.(map[string]interface{})["repository"].(map[string]interface{})
					results.ResositoriesWithFindings = append(results.ResositoriesWithFindings, models.RepoWithFindings{
						Nwo:   repoInfo["full_name"].(string),
						Count: int(repo.(map[string]interface{})["result_count"].(float64)),
						RunId: run.Id,
						Stars: int(repoInfo["stargazers_count"].(float64)),
					})
				}
			} else if repo.(map[string]interface{})["analysis_status"].(string) == "failed" {
				results.TotalFailedScans += 1
			}
		}

		skipped_repositories := runDetails["skipped_repositories"].(map[string]interface{})
		access_mismatch_repos := skipped_repositories["access_mismatch_repos"].(map[string]interface{})
		not_found_repos := skipped_repositories["not_found_repos"].(map[string]interface{})
		no_codeql_db_repos := skipped_repositories["no_codeql_db_repos"].(map[string]interface{})
		over_limit_repos := skipped_repositories["over_limit_repos"].(map[string]interface{})
		total_skipped_repos := access_mismatch_repos["repository_count"].(float64) + not_found_repos["repository_count"].(float64) + no_codeql_db_repos["repository_count"].(float64) + over_limit_repos["repository_count"].(float64)

		results.TotalSkippedAccessMismatchRepositories += int(access_mismatch_repos["repository_count"].(float64))
		results.TotalSkippedNotFoundRepositories += int(not_found_repos["repository_count"].(float64))
		results.TotalSkippedNoDatabaseRepositories += int(no_codeql_db_repos["repository_count"].(float64))
		results.TotalSkippedOverLimitRepositories += int(over_limit_repos["repository_count"].(float64))
		results.TotalSkippedRepositories += int(total_skipped_repos)
	}

	if jsonFlag {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Println("Run name:", sessionNameFlag)
		fmt.Println("Total runs:", len(results.Runs))
		fmt.Println("Total successful scans:", results.TotalSuccessfulScans)
		fmt.Println("Total failed scans:", results.TotalFailedScans)
		fmt.Println("Total skipped repositories:", results.TotalSkippedRepositories)
		fmt.Println("Total skipped repositories due to access mismatch:", results.TotalSkippedAccessMismatchRepositories)
		fmt.Println("Total skipped repositories due to not found:", results.TotalSkippedNotFoundRepositories)
		fmt.Println("Total skipped repositories due to no database:", results.TotalSkippedNoDatabaseRepositories)
		fmt.Println("Total skipped repositories due to over limit:", results.TotalSkippedOverLimitRepositories)
		fmt.Println("Total repositories with findings:", results.TotalRepositoriesWithFindings)
		fmt.Println("Total findings:", results.TotalFindingsCount)
		fmt.Println("Repositories with findings:")
		for _, repo := range results.ResositoriesWithFindings {
			fmt.Println("  ", repo.Nwo, ":", repo.Count)
		}
	}
}
