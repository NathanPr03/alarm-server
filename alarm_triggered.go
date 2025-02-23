package main

import (
	"go.uber.org/zap"
	"os/exec"
)

const soundAlarmScriptName = "sound_alarm.sh"
const triggerLightScriptName = "trigger_light.sh"

func WhenAlarmTriggered(contextLogger *zap.Logger, soundFileName string) {
	cmd := exec.Command("bash", soundAlarmScriptName, soundFileName)
	if err := cmd.Run(); err != nil {
		contextLogger.Error("Failed to run "+soundAlarmScriptName, zap.Error(err))
	} else {
		contextLogger.Info(soundAlarmScriptName + " executed successfully")
	}
}
