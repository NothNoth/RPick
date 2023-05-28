package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
)

type Resistor struct {
	Value     uint32
	Tolerance uint32
}

type Config struct {
	Resistors      []Resistor
	PopulationSize uint32
	MutationRate   uint32
}

type Individual struct {
	r1     Resistor
	r2     Resistor
	series bool

	value uint32
	note  float64
}

type ByNote []Individual

func (a ByNote) Len() int           { return len(a) }
func (a ByNote) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNote) Less(i, j int) bool { return a[i].note < a[j].note }

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], " <config file> <target value>")
		return
	}

	configFile := os.Args[1]
	target, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}

	configData, err := os.ReadFile(configFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(config)
	fmt.Println("Targeting", target, "Ohm")

	var population []Individual
	for i := 0; i < int(config.PopulationSize); i++ {
		population = append(population, generateIndividual(config.Resistors))
	}

	generation := 1
	for {
		//Evaluate
		for i := 0; i < int(config.PopulationSize); i++ {
			population[i].value, population[i].note = evaluate(&population[i], uint32(target))
		}

		//Sort
		sort.Sort(ByNote(population))

		//Select
		newPopulation := population[:config.PopulationSize/2.0]

		//Mix
		for i := 0; i < (int)(config.PopulationSize/2.0); i += 2 {
			newInd := mix(newPopulation[i], newPopulation[i+1], config.Resistors)

			//Mutate?
			if rand.Intn(100) < int(config.MutationRate) {
				newInd = mutate(newInd, config.Resistors)
			}
			newPopulation = append(newPopulation, newInd)
		}

		if generation%100 == 0 {
			fmt.Println("Generation", generation, " | ", population[0])
		}
		generation++
	}
}

func (r Resistor) String() string {
	return fmt.Sprintf("%dOhm (%d%%)", r.Value, r.Tolerance)
}

func (ind Individual) String() string {
	if ind.series {
		return fmt.Sprintf("%dOhm -- %dOhm => %d Ohm [%f]", ind.r1.Value, ind.r2.Value, ind.value, ind.note)
	}
	return fmt.Sprintf("%dOhm // %dOhm => %d Ohm [%f]", ind.r1.Value, ind.r2.Value, ind.value, ind.note)
}

func generateIndividual(resistors []Resistor) Individual {
	var ind Individual

	ind.r1 = resistors[rand.Intn(len(resistors))]
	ind.r2 = resistors[rand.Intn(len(resistors))]
	if rand.Intn(2) == 0 {
		ind.series = true
	} else {
		ind.series = false
	}

	return ind
}

func evaluate(ind *Individual, target uint32) (uint32, float64) {
	var value float64
	if ind.series {
		value = (float64)(ind.r1.Value + ind.r2.Value)
	} else {
		value = (1.0 / ((1.0 / (float64)(ind.r1.Value)) + (1.0 / (float64)(ind.r2.Value))))
	}

	return (uint32)(value), math.Abs((float64)(target) - value)
}

func mix(ind1 Individual, ind2 Individual, resistors []Resistor) Individual {
	var n Individual

	//Cannot mix different species, so create a new one
	if ind1.series != ind2.series {
		return generateIndividual(resistors)
	}

	if rand.Intn(2) == 0 {
		n.r1 = ind1.r1
		n.r2 = ind2.r2
	} else {
		n.r1 = ind1.r2
		n.r2 = ind2.r1
	}
	n.series = ind1.series

	return n
}

func mutate(ind Individual, resistors []Resistor) Individual {
	pick := resistors[rand.Intn(len(resistors))]

	if rand.Intn(2) == 0 {
		ind.r1 = pick
	} else {
		ind.r2 = pick
	}

	return ind
}
