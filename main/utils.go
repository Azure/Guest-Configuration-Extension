package main

import (
	"archive/zip"
	"fmt"
	"github.com/Azure/Guest-Configuration-Extension/pkg/seqnum"
	"github.com/mcuadros/go-version"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"encoding/json"
)

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// returns the filepaths for the unzipAgent and agent directories
func getAgentPaths() (unzipDirectory string, agentDirectory string) {
	unzipDirectory = filepath.Join(DataDir, UnzipAgentDir)
	agentDirectory = filepath.Join(unzipDirectory, AgentName)
	return unzipDirectory, agentDirectory
}

func parseAndLogAgentVersion(lg ExtensionLogger, agentName string) (agentVersion string, err error) {
	r, _ := regexp.Compile(AgentVersionRegex)
	matches := r.FindStringSubmatch(agentName)
	if len(matches) != 4 {
		return "", errors.New("incorrect naming format for agent")
	}
	// get the agent version
	agentVersion = matches[2]

	// logging and telemetry for agent version
	lg.customLog(logEvent, "current agent version", logVersion, agentVersion)
	telemetry(TelemetryScenario, "Current agent version: "+agentVersion, true, 0)

	return agentVersion, nil
}

func parseAndCompareExtensionVersions(lg ExtensionLogger, extensions []string) (extension string, err error) {
	r, _ := regexp.Compile(GCExtensionVersionRegex)

	var versions []string
	var match []string

	for _, ext := range extensions {
		match = r.FindStringSubmatch(ext)
		if len(match) != 3 {
			return "", errors.New("could not parse extension name from: " + ext)
		}
		versions = append(versions, match[2])
	}

	earliestVersion := versions[0]
	for _, v := range versions {
		if version.Compare(v, earliestVersion, "<") {
			earliestVersion = v
		}
	}

	lg.event("Found earliest version of the extension: " + earliestVersion)
	return match[1] + "-" + earliestVersion, nil
}

func getOldAgentPath(lg ExtensionLogger) (string, error) {
	// get the current path of the extension
	currentPath, err := os.Getwd()
	lg.event("Current path: " + currentPath)
	if err != nil {
		lg.eventError("failed to get current working directory path", err)
		return "", err
	}

	// get the directory of the current extension and read the files in it
	dir := filepath.Dir(currentPath)
	lg.event("Current Directory: " + dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		lg.eventError("could not read files in directory", err)
		return "", err
	}

	// get the two extensions in the directory
	var extensionDirs []string
	var matches bool
	for _, f := range files {
		matches, err = regexp.MatchString(ExtensionDirRegex, f.Name())
		if matches {
			extensionDirs = append(extensionDirs, f.Name())
		}
	}

	lg.event("All the extension directories found")
	for _, d := range extensionDirs {
		lg.event("dir: " + d)
	}

	// get the versions and compare them
	extension, err := parseAndCompareExtensionVersions(lg, extensionDirs)
	if err != nil {
		lg.eventError("failed to compare extension versions", err)
	}

	// get old agent path
	oldAgent := filepath.Join(dir, extension, UnzipAgentDir, AgentName)
	lg.event("Old agent path: " + oldAgent)

	return oldAgent, nil
}

// checkAndSaveSeqNum checks if the given seqNum is already processed
// according to the specified seqNumFile and if so, returns true,
// otherwise saves the given seqNum into seqNumFile returns false.
func checkAndSaveSeqNum(lg ExtensionLogger, seqNum int, mrseqPath string) (shouldExit bool, _ error) {
	lg.customLog(logEvent, "comparing seqnum", logPath, mrseqPath)
	smaller, err := seqnum.IsSmallerThan(mrseqPath, seqNum)
	if err != nil {
		return false, errors.Wrap(err, "failed to check sequence number")
	}
	if !smaller {
		// store sequence number is greater than the current sequence number
		return true, nil
	}
	if err := seqnum.Set(mrseqPath, seqNum); err != nil {
		return false, errors.Wrap(err, "failed to save the sequence number")
	}
	lg.customLog(logEvent, "seqNum saved", logPath, mrseqPath)

	return false, nil
}

// runCmd runs the command (extracted from cfg) in the given dir (assumed to exist).
func runCmd(lg ExtensionLogger, cmd string, dir string, cfg handlerSettings) (code int, err error) {
	lg.customLog(logEvent, "executing command", logOutput, dir)

	begin := time.Now()
	code, err = ExecCmdInDir(lg, cmd, dir)
	elapsed := time.Now().Sub(begin)
	isSuccess := err == nil

	lg.customLog(logEvent, "command executed", "command", cmd, "isSuccess", isSuccess, "time elapsed", elapsed)

	if err != nil {
		lg.customLog(logEvent, "failed to execute command", logError, err, logOutput, dir)
		return code, errors.Wrap(err, "failed to execute command")
	}
	lg.customLog(logEvent, "executed command", logOutput, dir)
	return code, nil
}

// decompresses a zip archive, moving all files and folders within the zip file
// to an output directory
func unzipAgent(lg ExtensionLogger, source string, prefix string, dest string) ([]string, error) {
	var filenames []string
	var agentZip = ""

	files, err := ioutil.ReadDir(source)
	if err != nil {
		return filenames, errors.Wrap(err, "failed to open the source dir: "+source)
	}

	for _, file := range files {
		if strings.Contains(file.Name(), prefix) {
			agentZip = filepath.Join(source, file.Name())
		}
	}

	if agentZip == "" {
		return filenames, errors.New("failed to find zip file " + agentZip)
	}

	lg.event("Got the agentZip. Agent is: " + agentZip)

	r, err := zip.OpenReader(agentZip)
	if err != nil {
		return filenames, errors.New("failed to open zip: " + agentZip)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, errors.Wrap(err, "failed to open file")
		}
		defer rc.Close()

		// store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// make folder
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			// make file
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, errors.Wrap(err, "failed to create directory: "+fpath)
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, errors.Wrap(err, "failed to open directory at current path: "+fpath)
			}
			_, err = io.Copy(outFile, rc)
			// close the file without defer to close before next iteration of loop
			outFile.Close()
			if err != nil {
				return filenames, errors.Wrap(err, "failed to close file: "+outFile.Name())
			}
		}
	}

	lg.event("unzipAgent successful")
	return filenames, nil
}

func setPermissions() error {
	_, agentDir := getAgentPaths()

	// get the list of files in the directory
	files, err := ioutil.ReadDir(agentDir)
	if err != nil {
		lg.eventError("could not read files in agent directory", err)
		return errors.Wrap(err, "could not read files in agent directory")
	}
	// set the permissions for each of the script files
	for _, f := range files {
		r, _ := regexp.Compile(".*\\.sh")
		matches := r.FindStringSubmatch(f.Name())
		if len(matches) > 0 {
			name := filepath.Join(agentDir, f.Name())
			err = os.Chmod(name, 0744)
			if err != nil {
				lg.eventError("could not set permissions for file: "+f.Name(), err)
				return errors.Wrap(err, "could not set permissions for file: "+f.Name())
			}
		}
	}

	return nil
}

func getStdPipesAndTelemetry(lg ExtensionLogger, logDir string, runErr error) {
	stdoutF, stderrF := logPaths(logDir)
	stdoutTail, err := tailFile(stdoutF, maxTailLen)
	if err != nil {
		lg.eventError("error tailing stdout logs", err)
	}
	stderrTail, err := tailFile(stderrF, maxTailLen)
	if err != nil {
		lg.eventError("error tailing stderr logs", err)
	}

	minStdout := min(len(stdoutTail), maxTelemetryTailLen)
	minStderr := min(len(stderrTail), maxTelemetryTailLen)
	msgTelemetry := fmt.Sprintf("\n[stdout]\n%s\n[stderr]\n%s",
		string(stdoutTail[len(stdoutTail)-minStdout:]),
		string(stderrTail[len(stderrTail)-minStderr:]))

	lg.event("Telemetry message: " + msgTelemetry)

	isSuccess := runErr == nil
	telemetry("output", msgTelemetry, isSuccess, 0)
}

func updateAssignment(assignmentName string, contentHash string) {
    if( assignmentName == "" || contentHash == "" ) {
        return
    }

    dscConfigFilePath := "/var/lib/GuestConfig/dsc/dsc.config"
    fileMode := int(0644)

    configJsonFile, err := os.Open(dscConfigFilePath)
    if err != nil {
        dscConfigFolderPath := "/var/lib/GuestConfig/dsc"
        os.MkdirAll(dscConfigFolderPath, os.ModePerm)

        // New dsc.config file
        firstAssignmentString := "{\"Assignments\": [ { \"name\": \"" + assignmentName + "\", \"contentHash\": \"" + contentHash + "\" }]}"
        var firstAssignment map[string]interface{}
        json.Unmarshal([]byte(firstAssignmentString ), &firstAssignment)
        firstAssignmentJson, _ := json.Marshal(firstAssignment)
        err = ioutil.WriteFile(dscConfigFilePath, firstAssignmentJson, fileMode)
        if err != nil {
            return errors.Wrap(err, "failed to open file")
        }
        return nil
    }
    
    // defer the closing of config file so that we can parse it later on
    defer configJsonFile.Close()

    byteValue, _ := ioutil.ReadAll(configJsonFile)

    var dscConfig map[string]interface{}
    json.Unmarshal([]byte(byteValue), &dscConfig)

    newEntry := map[string]interface{}{"name" : assignmentName, "contentHash" : contentHash}
    
    // Assignments exists in config file
    if _, ok := dscConfig["Assignments"]; ok {
        assignments := dscConfig["Assignments"].([]interface{})
        
        var assignmentExists bool = false
        for _, assignment := range assignments {
            assignmentMap := assignment.(map[string]interface{})
            if assignmentMap["name"] == assignmentName {
                assignmentExists = true
                assignmentMap["contentHash"] = contentHash
            }
        }
        
        if assignmentExists == false {
            dscConfig["Assignments"] = append(assignments, newEntry)
        }

    } else {
        // Assignments doesnt exist in config file
        var assignments []interface{}
        dscConfig["Assignments"] = append(assignments, newEntry)
    }

    dscConfigJson, _ := json.Marshal(dscConfig)
    err = ioutil.WriteFile(dscConfigFilePath, dscConfigJson, fileMode)
    if err != nil {
        return errors.Wrap(err, "failed to write file")
    }

    return nil
}