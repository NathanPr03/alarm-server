package main

import (
	"go.uber.org/zap"
	"os/exec"
)

func WhenLightTriggered(contextLogger *zap.Logger) {
	cmd := exec.Command("bash", triggerLightScriptName)
	if err := cmd.Run(); err != nil {
		contextLogger.Error("Failed to run "+triggerLightScriptName, zap.Error(err))
	} else {
		contextLogger.Info(triggerLightScriptName + " executed successfully")
	}
}
