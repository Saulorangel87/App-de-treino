package activityimport

import "time"

// bestPowerWindow é a janela do teste de FTP de 20 minutos.
const bestPowerWindow = 20 * 60

// maxPowerGapSeconds é a maior pausa entre dois registros que ainda conta como
// "mantendo a mesma potência". Aparelhos com gravação inteligente pulam
// segundos; uma pausa maior (parada, arquivo cortado) entra como zero watt, o
// que impede de inventar uma janela forte onde o atleta estava parado.
const maxPowerGapSeconds = 10

type powerSample struct {
	at    time.Time
	watts int
}

// bestAveragePower devolve a maior potência média de uma janela contínua de
// `window` segundos, ou nil quando o arquivo não tem potência ou é mais curto
// que a janela. Os registros viram uma série de 1 Hz (mantém o último valor em
// lacunas curtas e usa zero nas longas) antes de percorrer a janela.
func bestAveragePower(samples []powerSample, window int) *int {
	if len(samples) < 2 || window <= 0 {
		return nil
	}
	start := samples[0].at
	total := int(samples[len(samples)-1].at.Sub(start).Seconds()) + 1
	if total < window || total > 24*3600 {
		return nil
	}
	series := make([]int, total)
	for i, sample := range samples {
		from := int(sample.at.Sub(start).Seconds())
		if from < 0 || from >= total {
			continue
		}
		series[from] = sample.watts
		if i+1 < len(samples) {
			next := int(samples[i+1].at.Sub(start).Seconds())
			for second := from + 1; second < next && second-from <= maxPowerGapSeconds; second++ {
				series[second] = sample.watts
			}
		}
	}
	sum, best := 0, 0
	for i, watts := range series {
		sum += watts
		if i >= window {
			sum -= series[i-window]
		}
		if i >= window-1 && sum > best {
			best = sum
		}
	}
	if best == 0 {
		return nil
	}
	value := best / window
	if value == 0 {
		return nil
	}
	return &value
}
