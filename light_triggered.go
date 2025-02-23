package main

import (
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
)

const logFilePath = "alarm_sh.log"

func WhenLightTriggered(contextLogger *zap.Logger) {
	cmd := exec.Command("bash", triggerLightScriptName)
	if err := cmd.Run(); err != nil {
		contextLogger.Error("Failed to run "+triggerLightScriptName, zap.Error(err))
		logContents, readErr := readLogFile(logFilePath)
		if readErr != nil {
			contextLogger.Error("Failed to read log file", zap.Error(readErr))
		} else {
			contextLogger.Error("Log file contents:\n" + logContents)
		}
	} else {
		contextLogger.Info(triggerLightScriptName + " executed successfully")
	}
}

func readLogFile(filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("log file %s does not exist", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %v", err)
	}

	return string(data), nil
}
