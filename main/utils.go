package main

import (
	"archive/zip"
	"fmt"
	"github.com/mcuadros/go-version"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/Guest-Configuration-Extension/pkg/seqnum"
)

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// returns the filepaths for the unzip and agent directories
func unzipAndAgentDirectories() (unzipDirectory string, agentDirectory string) {
	unzipDirectory = filepath.Join(dataDir, agentDir)
	agentDirectory = filepath.Join(unzipDirectory, agentName)
	return unzipDirectory, agentDirectory
}

func parseAndLogAgentVersion(agentName string) (agentVersion string, err error) {
	r, _ := regexp.Compile("^([./a-zA-Z0-9]*)_([0-9.]*)?[.](.*)$")
	matches := r.FindStringSubmatch(agentName)
	if len(matches) != 4 {
		return "", errors.New("incorrect naming format for agent")
	}
	// get the agent version
	agentVersion = matches[2]

	// logging and telemetry for agent version
	lg.customLog(logMessage, "current agent version", logVersion, agentVersion)
	telemetry(telemetryScenario, "Current agent version: "+agentVersion, true, 0)

	return agentVersion, nil
}

func parseAndCompareExtensionVersions(extension1 string, extension2 string) (extension string, err error) {
	r, _ := regexp.Compile("^([./a-zA-Z]*)-([0-9.]*)?$")
	matches := r.FindStringSubmatch(extension1)
	if len(matches) != 3 {
		return "", errors.New("could not parse extension name")
	}
	version1 := matches[2]

	matches = r.FindStringSubmatch(extension2)
	if len(matches) != 3 {
		return "", errors.New("could not parse extension name")
	}
	version2 := matches[2]

	// compare versions
	v1smaller := version.Compare(version1, version2, "<")
	if v1smaller == true {
		return extension1, nil
	} else {
		return extension2, nil
	}
}

func getOldAgentPath() (string, error) {
	// get the current path of the extension
	currentPath, err := os.Getwd()
	if err != nil {
		lg.messageAndError("failed to get current working directory path", err)
		return "", err
	}

	// get the directory of the current extension and read the files in it
	dir := filepath.Dir(currentPath)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		lg.messageAndError("could not read files in directory", err)
		return "", err
	}

	// get the two extensions in the directory
	var extensionDirs [2]string
	i := 0
	for _, f := range files {
		if strings.Contains(f.Name(), "Microsoft.GuestConfiguration.Edp.ConfigurationForLinux") {
			extensionDirs[i] = f.Name()
			i++
		}
		if len(extensionDirs) <= i {
			break
		}
	}

	// get the versions and compare them
	extension, err := parseAndCompareExtensionVersions(extensionDirs[0], extensionDirs[1])
	if err != nil {
		lg.messageAndError("failed to compare extension versions", err)
	}

	// get old agent path
	oldAgent := filepath.Join(dir, extension, agentDir, agentName)

	return oldAgent, nil
}

// checkAndSaveSeqNum checks if the given seqNum is already processed
// according to the specified seqNumFile and if so, returns true,
// otherwise saves the given seqNum into seqNumFile returns false.
func checkAndSaveSeqNum(seqNum int, mrseqPath string) (shouldExit bool, _ error) {
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
	lg.customLog(logMessage, "seqNum saved", logPath, mrseqPath)

	return false, nil
}

// runCmd runs the command (extracted from cfg) in the given dir (assumed to exist).
func runCmd(cmd string, dir string, cfg handlerSettings) (err error) {
	lg.customLog(logEvent, "executing command", logOutput, dir)

	begin := time.Now()
	err = ExecCmdInDir(cmd, dir)
	elapsed := time.Now().Sub(begin)
	isSuccess := err == nil

	lg.customLog(logMessage, "command executed", "command", cmd, "isSuccess", isSuccess, "time elapsed", elapsed)

	if err != nil {
		lg.customLog(logMessage, "failed to execute command", logError, err, logOutput, dir)
		return errors.Wrap(err, "failed to execute command")
	}
	lg.customLog(logEvent, "executed command", logOutput, dir)
	return nil
}

// decompresses a zip archive, moving all files and folders within the zip file
// to an output directory
func unzip(source string, dest string) ([]string, error) {
	lg.event("begin unzipping agent", "")
	var filenames []string
	r, err := zip.OpenReader(source)
	if err != nil {
		return filenames, errors.Wrap(err, "failed to open zip")
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
				return filenames, errors.Wrap(err, "failed to create directory")
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, errors.Wrap(err, "failed to open directory at current path")
			}
			_, err = io.Copy(outFile, rc)
			// close the file without defer to close before next iteration of loop
			outFile.Close()
			if err != nil {
				return filenames, errors.Wrap(err, "failed to close file")
			}
		}
	}
	lg.message("unzip successful")
	return filenames, nil
}

func getStdPipesAndTelemetry(logDir string, runErr error) {
	stdoutF, stderrF := logPaths(logDir)
	stdoutTail, err := tailFile(stdoutF, maxTailLen)
	if err != nil {
		lg.messageAndError("error tailing stdout logs", err)
	}
	stderrTail, err := tailFile(stderrF, maxTailLen)
	if err != nil {
		lg.messageAndError("error tailing stderr logs", err)
	}

	minStdout := min(len(stdoutTail), maxTelemetryTailLen)
	minStderr := min(len(stderrTail), maxTelemetryTailLen)
	msgTelemetry := fmt.Sprintf("\n[stdout]\n%s\n[stderr]\n%s",
		string(stdoutTail[len(stdoutTail)-minStdout:]),
		string(stderrTail[len(stderrTail)-minStderr:]))

	isSuccess := runErr == nil
	telemetry("output", msgTelemetry, isSuccess, 0)
}
