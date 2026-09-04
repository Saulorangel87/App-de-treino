package planning

// sessionProtocol is the reviewed prescription template selected by the
// rules engine. The template controls the shape of a session; the engine still
// applies the athlete's duration, level and safety limits before persisting it.
type sessionProtocol struct {
	Key             string
	EvidenceKeys    []string
	EvidenceScope   string
	WorkMinutes     int
	RecoveryMinutes int
	Repetitions     int
	WorkTitle       string
	WorkInstruction string
}

var sessionProtocols = map[string]sessionProtocol{
	"Giro de base": {
		Key: "base_endurance", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
	},
	"Endurance contínuo": {
		Key: "continuous_endurance", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
	},
	"Giro leve protegido": {
		Key: "protected_recovery", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Redução conservadora de carga por segurança; não substitui avaliação profissional.",
	},
	"Tempo controlado": {
		Key: "controlled_tempo", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
		WorkMinutes:   8, RecoveryMinutes: 3, Repetitions: 3,
		WorkTitle: "Ritmo controlado", WorkInstruction: "Sustente um ritmo estável em que ainda consiga manter a técnica e a respiração sob controle.",
	},
	"Ritmo de prova controlado": {
		Key: "controlled_event_pace", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
		WorkMinutes:   8, RecoveryMinutes: 3, Repetitions: 3,
		WorkTitle: "Ritmo controlado", WorkInstruction: "Sustente um ritmo estável em que ainda consiga manter a técnica e a respiração sob controle.",
	},
	"Cadência técnica": {
		Key: "technical_cadence", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
		WorkMinutes:   2, RecoveryMinutes: 2, Repetitions: 6,
		WorkTitle: "Cadência técnica", WorkInstruction: "Aumente a cadência mantendo o movimento redondo e sem tensão excessiva.",
	},
	"Subidas controladas": {
		Key: "controlled_hills", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
		WorkMinutes:   6, RecoveryMinutes: 3, Repetitions: 4,
		WorkTitle: "Subida controlada", WorkInstruction: "Mantenha um esforço firme e controlado na subida, sem sprintar.",
	},
	"Sweet spot por potência": {
		Key: "power_sweet_spot", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; o FTP informado orienta o contexto, não uma meta universal.",
		WorkMinutes:   8, RecoveryMinutes: 4, Repetitions: 3,
		WorkTitle: "Bloco sustentável", WorkInstruction: "Mantenha um esforço forte e sustentável, sem transformar o bloco em um sprint.",
	},
	"Sweet spot progressivo": {
		Key: "progressive_sweet_spot", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
		WorkMinutes:   8, RecoveryMinutes: 4, Repetitions: 3,
		WorkTitle: "Bloco sustentável", WorkInstruction: "Mantenha um esforço forte e sustentável, sem transformar o bloco em um sprint.",
	},
	"Intervalos controlados": {
		Key: "controlled_intervals", EvidenceKeys: []string{"acsm-1998", "rosenblat-2020"},
		EvidenceScope: "A revisão específica informa a escolha de intervalos, sem tratar seus tempos como uma prescrição universal.",
		WorkMinutes:   4, RecoveryMinutes: 3, Repetitions: 4,
		WorkTitle: "Intervalo forte-controlado", WorkInstruction: "Sustente o esforço forte com técnica estável; reduza o ritmo se perder o controle.",
	},
	"Intervalos moderados de estrada": {
		Key: "road_moderate_intervals", EvidenceKeys: []string{"road-mit-block-2025", "road-intensity-2024"},
		EvidenceScope: "A evidência apoia estímulos intervalados moderados em ciclistas treinados e compara distribuições de intensidade, mas não valida esta dose para todos os perfis. A sessão é uma adaptação conservadora, não a reprodução dos blocos estudados.",
		WorkMinutes:   10, RecoveryMinutes: 3, Repetitions: 3,
		WorkTitle: "Intervalo moderado", WorkInstruction: "Sustente um esforço moderado e firme, mantendo a técnica e terminando o bloco sem sprintar; reduza o ritmo se perder o controle.",
	},
}

func protocolForWorkout(name string) sessionProtocol {
	if protocol, ok := sessionProtocols[name]; ok {
		return protocol
	}
	return sessionProtocol{
		Key: "continuous_base", EvidenceKeys: []string{"acsm-1998"},
		EvidenceScope: "Progressão gradual e controle de carga; a referência não define minutos universais.",
	}
}
