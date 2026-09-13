package planning

const qualitySessionDensityGateThreshold = 2

func stimulusDistributionRequiresRecovery(distribution TrainingStimulusDistribution) bool {
	return distribution.QualitySessionsLast7d >= qualitySessionDensityGateThreshold ||
		distribution.AdjacentQualitySessionPairs > 0
}

func stimulusDistributionHasIncompleteData(distribution TrainingStimulusDistribution) bool {
	return len(distribution.MissingData) > 0 || len(distribution.DataIssues) > 0
}

func stimulusDistributionGateTriggered(distribution TrainingStimulusDistribution) bool {
	return stimulusDistributionHasIncompleteData(distribution) || stimulusDistributionRequiresRecovery(distribution)
}

func appendStimulusDistributionQualityReasons(addReason func(string, string), distribution TrainingStimulusDistribution) bool {
	triggered := false
	if distribution.AdjacentQualitySessionPairs > 0 {
		addReason("recent_quality_session_proximity", "Foram observadas sessões de qualidade em dias consecutivos no histórico de 42 dias; a carga não deve ser tratada como elegível para progressão.")
		triggered = true
	}
	if distribution.QualitySessionsLast7d >= qualitySessionDensityGateThreshold {
		addReason("recent_quality_session_density", "Foram observadas pelo menos duas sessões de qualidade nos últimos 7 dias; a densidade requer observação antes de considerar progressão.")
		triggered = true
	}
	return triggered
}

func appendStimulusDistributionQuality(missing *[]string, dataIssues *[]string, distribution TrainingStimulusDistribution) {
	for _, value := range distribution.MissingData {
		*missing = appendUniqueString(*missing, value)
	}
	for _, value := range distribution.DataIssues {
		*dataIssues = appendUniqueString(*dataIssues, value)
	}
}
