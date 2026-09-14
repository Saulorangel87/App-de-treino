package planning

// ProtocolMetadata keeps the operational limits of each workout template in a
// versioned, testable structure. Duration, RPE and session steps remain in the
// generated Workout and its Structure.
type ProtocolMetadata struct {
	PhysiologicalObjective string   `json:"physiological_objective"`
	PracticalObjective     string   `json:"practical_objective"`
	Indication             string   `json:"indication"`
	Contraindication       string   `json:"contraindication"`
	RecommendedLevel       string   `json:"recommended_level"`
	Prerequisites          []string `json:"prerequisites"`
	HeartRateGuidance      string   `json:"heart_rate_guidance"`
	PowerGuidance          string   `json:"power_guidance"`
	CadenceGuidance        string   `json:"cadence_guidance"`
	StopCriteria           string   `json:"stop_criteria"`
	ProgressionCriteria    string   `json:"progression_criteria"`
	RegressionCriteria     string   `json:"regression_criteria"`
}

func metadataForProtocol(key string) ProtocolMetadata {
	common := ProtocolMetadata{
		Contraindication:    "Dor, limitação ativa, restrição médica incompatível ou necessidade de recuperação prevalecem sobre o estímulo.",
		Prerequisites:       []string{"disponibilidade compatível", "dados mínimos válidos", "ausência de sinal protetivo"},
		HeartRateGuidance:   "Opcional e apenas observacional; não substitui RPE nem avaliação individual.",
		PowerGuidance:       "Opcional e apenas observacional, salvo quando o protocolo exigir FTP informado.",
		CadenceGuidance:     "Livre e confortável, salvo quando o objetivo técnico declarar outra orientação.",
		StopCriteria:        "Interromper ou reduzir diante de dor, tontura, mal-estar, falta de ar incomum ou perda de controle técnico.",
		ProgressionCriteria: "Somente após conclusão coerente, recuperação adequada e nova avaliação determinística do plano.",
		RegressionCriteria:  "Reduzir duração ou RPE quando houver dificuldade excessiva, fadiga, dor, recuperação insuficiente ou dados inconsistentes.",
	}

	switch key {
	case "protected_recovery":
		common.PhysiologicalObjective = "Preservar movimento com carga interna mínima."
		common.PracticalObjective = "Oferecer uma sessão protegida sem estímulo de qualidade."
		common.Indication = "Sinal protetivo, limitação ou recuperação insuficiente identificado pelas regras."
		common.RecommendedLevel = "todos os níveis"
		common.ProgressionCriteria = "Não progride por si só; exige reavaliação posterior sem sinais protetivos."
	case "active_recovery":
		common.PhysiologicalObjective = "Favorecer recuperação mantendo atividade aeróbica muito leve."
		common.PracticalObjective = "Manter o hábito na semana de recuperação sem acrescentar qualidade."
		common.Indication = "Semana programada de recuperação, sem limitação ou dor ativa."
		common.RecommendedLevel = "todos os níveis"
		common.ProgressionCriteria = "Retomar o ciclo regular somente após recuperação adequada."
	case "return_after_break":
		common.PhysiologicalObjective = "Readaptar gradualmente à carga aeróbica após uma pausa."
		common.PracticalObjective = "Retomar consistência com duração e esforço limitados."
		common.Indication = "Atleta que declarou retorno após pausa."
		common.RecommendedLevel = "todos os níveis"
		common.ProgressionCriteria = "Sair da retomada apenas após atualização explícita da situação de treino e boa tolerância."
	case "base_endurance", "continuous_endurance", "continuous_base":
		common.PhysiologicalObjective = "Desenvolver ou manter capacidade aeróbica em esforço sustentável."
		common.PracticalObjective = "Acumular tempo de pedal com ritmo estável e controlado."
		common.Indication = "Sessão geral de base compatível com o nível e a disponibilidade."
		common.RecommendedLevel = "todos os níveis"
	case "long_endurance":
		common.PhysiologicalObjective = "Desenvolver tolerância a um esforço aeróbico contínuo mais prolongado."
		common.PracticalObjective = "Usar o maior slot disponível sem criar meta universal de distância."
		common.Indication = "Maior disponibilidade semanal, sem sinal protetivo e dentro do limite do nível."
		common.RecommendedLevel = "todos os níveis"
	case "technical_cadence":
		common.PhysiologicalObjective = "Praticar coordenação do gesto de pedalada sob carga controlada."
		common.PracticalObjective = "Aumentar a cadência sem perder fluidez ou técnica."
		common.Indication = "Preferência por cadência ou contexto indoor elegível."
		common.RecommendedLevel = "intermediário e avançado"
		common.CadenceGuidance = "Elevar gradualmente a cadência sem meta universal de RPM e sem tensão excessiva."
	case "controlled_hills":
		common.PhysiologicalObjective = "Desenvolver capacidade aeróbica e tolerância a esforços sustentados em subida."
		common.PracticalObjective = "Praticar subidas sem sprint ou esforço máximo."
		common.Indication = "Terreno com subidas informado e atleta elegível."
		common.RecommendedLevel = "intermediário e avançado"
		common.CadenceGuidance = "Autorregulada conforme inclinação e relação disponível, sem impor baixa cadência."
	case "power_sweet_spot":
		common.PhysiologicalObjective = "Sustentar esforço submáximo de qualidade com controle de carga."
		common.PracticalObjective = "Executar blocos sustentáveis usando o FTP informado como contexto."
		common.Indication = "Atleta avançado com medidor de potência e FTP informado."
		common.RecommendedLevel = "avançado"
		common.Prerequisites = append(common.Prerequisites, "medidor de potência", "FTP informado")
		common.PowerGuidance = "Usar o FTP informado apenas como orientação contextual; RPE e sinais de segurança prevalecem."
	case "progressive_sweet_spot", "controlled_tempo":
		common.PhysiologicalObjective = "Desenvolver capacidade de sustentar esforço aeróbico moderado a forte."
		common.PracticalObjective = "Completar blocos estáveis sem sprintar ou perder controle."
		common.Indication = "Sessão de qualidade compatível com nível, fase e recuperação."
		common.RecommendedLevel = "avançado"
	case "controlled_event_pace":
		common.PhysiologicalObjective = "Praticar esforço sustentável relacionado à demanda do evento."
		common.PracticalObjective = "Ensaiar ritmo percebido sem simular a prova inteira."
		common.Indication = "Atleta avançado, evento futuro válido, avaliação apta e fase específica."
		common.RecommendedLevel = "avançado"
		common.Prerequisites = append(common.Prerequisites, "evento futuro válido", "avaliação submáxima apta")
	case "controlled_threshold":
		common.PhysiologicalObjective = "Desenvolver tolerância a esforço forte e estável próximo ao limiar percebido."
		common.PracticalObjective = "Executar blocos controlados sem estimar limiar ou impor potência."
		common.Indication = "Estrada ou indoor, nível avançado, histórico mínimo, objetivo compatível e avaliação apta."
		common.RecommendedLevel = "avançado"
		common.Prerequisites = append(common.Prerequisites, "avaliação submáxima apta", "histórico mínimo")
	case "controlled_intervals", "road_moderate_intervals":
		common.PhysiologicalObjective = "Desenvolver capacidade aeróbica em esforços intervalados controlados."
		common.PracticalObjective = "Alternar blocos de trabalho e recuperação mantendo técnica estável."
		common.Indication = "Objetivo de performance ou evento, avaliação apta e semana de construção."
		common.RecommendedLevel = "intermediário e avançado conforme o protocolo"
		common.Prerequisites = append(common.Prerequisites, "avaliação submáxima apta")
	case "road_high_intensity_intervals", "road_vo2_intervals", "short_self_regulated_intervals", "xco_aerobic_intervals":
		common.PhysiologicalObjective = "Desenvolver capacidade aeróbica de alta intensidade com recuperação estruturada."
		common.PracticalObjective = "Executar intervalos fortes controlados, sem sprint máximo."
		common.Indication = "Atleta avançado na modalidade elegível, com histórico, avaliação e disponibilidade mínimos."
		common.RecommendedLevel = "avançado"
		common.Prerequisites = append(common.Prerequisites, "avaliação submáxima apta", "histórico mínimo", "modalidade elegível")
	default:
		common.PhysiologicalObjective = "Desenvolver capacidade aeróbica com carga controlada."
		common.PracticalObjective = "Realizar uma sessão contínua compatível com o contexto disponível."
		common.Indication = "Fallback conservador quando não há protocolo específico elegível."
		common.RecommendedLevel = "todos os níveis"
	}

	return common
}
