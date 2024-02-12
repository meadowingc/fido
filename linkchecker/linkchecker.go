package linkchecker

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"strings"
)

type LinkCheckResult struct {
	Summary     string
	FoundErrors []LinkCheckError
}

type LinkCheckError struct {
	Name           string
	ParentURL      string
	RealURL        string
	CheckTime      string
	Warning        string
	CheckingResult string
}

func getAnalysisValueForKey(key string, line string) string {
	components := strings.Split(line, key)
	value := strings.Join(components[1:], " ")
	value = strings.TrimSpace(value)
	return value
}

func CheckLink(link string) (LinkCheckResult, error) {
	// linkchecker https://meadow.bearblog.dev -o csv --no-status --no-warnings --timeout 1200
	cmd := exec.Command("linkchecker", link, "-o", "text", "--no-status", "--no-warnings", "--timeout", "1200")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {

		if exitError, ok := err.(*exec.ExitError); ok {
			// if exit code is 0 then we can return early since there were no issues
			if exitError.ExitCode() == 0 {
				return LinkCheckResult{
					Summary: "No issues found",
				}, nil
			}

			if exitError.ExitCode() == 2 {
				return LinkCheckResult{}, errors.New("there was an error while checking for links")
			}

			// exit code of 1 is expected IF there were any issues
		} else {
			return LinkCheckResult{}, errors.New("there was an UNKNOWN error while checking for links")
		}
	}

	outStr := out.String()

	resultsChunk := strings.Split(outStr, "Start checking at")[1]
	resultsChunk = strings.Split(resultsChunk, "Stopped checking at")[0]

	summaryLine := strings.Split(resultsChunk, "That's it. ")[1]
	summaryLine = strings.TrimSpace(summaryLine)

	analysisResultsChunk := strings.Split(resultsChunk, "Statistics:")[0]
	analysisResults := strings.TrimSpace(analysisResultsChunk)

	resultsArray := strings.Split(analysisResults, "\n\n")

	allCheckResults := []LinkCheckError{}

	for _, result := range resultsArray {
		linkCheckError := LinkCheckError{}

		for _, line := range strings.Split(result, "\n") {
			if strings.HasPrefix(line, "Real URL") {
				linkCheckError.RealURL = getAnalysisValueForKey("URL", line)
			} else if strings.HasPrefix(line, "Name") {
				linkCheckError.Name = getAnalysisValueForKey("Name", line)
				linkCheckError.Name = strings.TrimLeft(linkCheckError.Name, "`")
				linkCheckError.Name = strings.TrimRight(linkCheckError.Name, "'")
			} else if strings.HasPrefix(line, "Parent URL") {
				linkCheckError.ParentURL = getAnalysisValueForKey("Parent URL", line)
				linkCheckError.ParentURL = strings.Split(linkCheckError.ParentURL, ",")[0]
			} else if strings.HasPrefix(line, "Check time") {
				linkCheckError.CheckTime = getAnalysisValueForKey("Check time", line)
			} else if strings.HasPrefix(line, "Warning") {
				linkCheckError.Warning = getAnalysisValueForKey("Warning", line)
			} else if strings.HasPrefix(line, "          ") {
				linkCheckError.Warning += " " + strings.TrimSpace(line)
			} else if strings.HasPrefix(line, "Result") {
				linkCheckError.CheckingResult = getAnalysisValueForKey("Result", line)
			} else {
				log.Println("Unknown line: ", line)
			}
		}

		if linkCheckError.Name != "" {
			allCheckResults = append(allCheckResults, linkCheckError)
		}
	}

	return LinkCheckResult{
		Summary:     summaryLine,
		FoundErrors: allCheckResults,
	}, nil
}
