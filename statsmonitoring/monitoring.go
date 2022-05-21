package statsmonitoring

import (
	"log"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

type Metrics struct {
	CurrCPUUsagePercentage      float32
	ThirtySecondCPUUsageAverage float32
	SixtySecondCPUUsageAverage  float32

	CurrRamUsagePercentage                float32
	ThirtySecondRamUsagePercentageAverage float32
	SixtySecondRamUsagePercentageAverage  float32
}

var (
	cpuUsagePercentages    []float32
	memoryUsagePercentages []float32
)

func init() {
	for i := 0; i < 60; i++ {
		cpuUsagePercentages = append(cpuUsagePercentages, 0)
		memoryUsagePercentages = append(memoryUsagePercentages, 0)
	}
}

func getTotalRAM() float32 {
	memory, err := memory.Get()
	if err != nil {
		log.Fatal(err)
	}
	return float32(((memory.Total / 1000) / 1000))
}

func getCPUUsagePercentage() {
	for {
		before, err := cpu.Get()
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Second)
		after, err := cpu.Get()
		if err != nil {
			log.Fatal(err)
		}
		total := float64(after.Total - before.Total)

		currCPUUsage := float32(float64(after.User-before.User) / total * 100)
		cpuUsagePercentages = append([]float32{currCPUUsage}, cpuUsagePercentages...)
		if len(cpuUsagePercentages) > 6 {
			cpuUsagePercentages = cpuUsagePercentages[:6]
		}
		time.Sleep(1 * time.Second)
	}
}

func GetCurrentRAMUsagePercentage() {
	totalRAM := getTotalRAM()
	for {
		memory, err := memory.Get()
		if err != nil {
			log.Fatal(err)
		}

		currMemoryUsage := 100 * float32(float64(((memory.Used / 1000) / 1000)))
		currMemoryUsagePercentage := float32(currMemoryUsage / totalRAM)

		memoryUsagePercentages = append([]float32{currMemoryUsagePercentage}, memoryUsagePercentages...)
		if len(memoryUsagePercentages) > 6 {
			memoryUsagePercentages = memoryUsagePercentages[:6]
		}
		time.Sleep(1 * time.Second)
	}
}

func getAverageOfNumbers(slice []float32) float32 {
	var sum float32
	for _, val := range slice {
		sum += float32(val)
	}
	return sum / float32(len(slice))
}

func GetMetrics() Metrics {
	return Metrics{
		CurrCPUUsagePercentage:                cpuUsagePercentages[0],
		ThirtySecondCPUUsageAverage:           getAverageOfNumbers(cpuUsagePercentages[:3]),
		SixtySecondCPUUsageAverage:            getAverageOfNumbers(cpuUsagePercentages),
		CurrRamUsagePercentage:                memoryUsagePercentages[0],
		ThirtySecondRamUsagePercentageAverage: getAverageOfNumbers(memoryUsagePercentages[:3]),
		SixtySecondRamUsagePercentageAverage:  getAverageOfNumbers(memoryUsagePercentages),
	}
}

func CollectAndShipStats() {
	go getCPUUsagePercentage()
	go GetCurrentRAMUsagePercentage()
}
