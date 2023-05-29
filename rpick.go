package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
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

type Mode int

/*
		  Serial : --[   ] - [  ]--
			Parallel:   -[    ]-
			           --[    ]--
		  Serial3 : --[   ] - [   ] - [   ]--
			Parallel3:   -[    ]-
			            --[    ]--
			             -[    ]-
	    Combo:			-[    ]-
			           --[    ]---[   ]--
*/
const (
	eModeSerial           Mode = iota
	eModeParallel         Mode = iota
	eModeSerial3          Mode = iota
	eModeParallel3        Mode = iota
	eModeParallelAndSerie Mode = iota
	eModeSeriesOnParallel Mode = iota
	eModeMax              Mode = iota
)

type Individual struct {
	r1   Resistor
	r2   Resistor
	r3   Resistor //Optional
	mode Mode

	value     uint32
	tolerance uint32
	note      float64
}

type ByNote []Individual

func (a ByNote) Len() int           { return len(a) }
func (a ByNote) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNote) Less(i, j int) bool { return a[i].note < a[j].note }

func main() {
	var duplicates uint32
	var config Config
	var population []Individual

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

	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(config)
	fmt.Println("Targeting", target, "Ohm")

	kill := false
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			kill = true
			fmt.Println(sig)
		}
	}()

	for i := 0; i < int(config.PopulationSize); i++ {
		population = append(population, generateIndividual(config.Resistors))
	}

	generation := 1
	mutationRate := config.MutationRate
	for {
		//Evaluate
		for i := 0; i < int(config.PopulationSize); i++ {
			population[i].value, population[i].tolerance, population[i].note = evaluate(&population[i], uint32(target))
		}

		//Sort
		sort.Sort(ByNote(population))
		/*
			//Debug
			fmt.Printf("Currently at generation %d [Mrate: %d]\n", generation, mutationRate)
			for idx, ind := range population {
				fmt.Println(idx, "", ind)
			}
			// /Debug
		*/
		if kill {
			break
		}

		//Select first half
		newPopulation := population[:config.PopulationSize/2.0]

		//Mix consecutives from first half to generate second half
		for i := 0; i < (int)(config.PopulationSize/2.0); i += 2 {
			newInd := mix(newPopulation[i], newPopulation[i+1], config.Resistors)

			//Mutate?
			if rand.Intn(100) < int(mutationRate) {
				newInd = mutate(newInd, config.Resistors)
			}
			newPopulation = append(newPopulation, newInd)
		}

		population = newPopulation[:config.PopulationSize]

		if generation%100 == 0 {
			duplicates, population = cleanup(population)
			duplicatesRatio := float64(duplicates) * 100.0 / float64(config.PopulationSize)
			fmt.Println("Generation", generation, "(duplicates", duplicatesRatio, "%) | ", population[0])
			if duplicatesRatio > 30 {
				mutationRate = config.MutationRate * 2.0
			} else {
				mutationRate = config.MutationRate
			}
		}
		generation++
		//time.Sleep(1 * time.Millisecond)
	}

	duplicates, population = cleanup(population)
	fmt.Printf("Stopped at generation %d [%.0f %% duplicates]", generation, float64(duplicates)*100.0/float64(config.PopulationSize))
	for idx, ind := range population {
		fmt.Println(idx, "", ind)
	}
	fmt.Println("Best result is:", population[0])

}

func (r Resistor) String() string {
	return fmt.Sprintf("%d Ohm (%d%%)", r.Value, r.Tolerance)
}

func (ind Individual) String() string {
	switch ind.mode {
	case eModeSerial:
		return fmt.Sprintf("%d Ohm -- %d Ohm => %d Ohm %d%% [%f]", ind.r1.Value, ind.r2.Value, ind.value, ind.tolerance, ind.note)
	case eModeSerial3:
		return fmt.Sprintf("%d Ohm -- %d Ohm -- %d Ohm => %d Ohm %d%% [%f]", ind.r1.Value, ind.r2.Value, ind.r3.Value, ind.value, ind.tolerance, ind.note)
	case eModeParallel:
		return fmt.Sprintf("%d Ohm // %d Ohm => %d Ohm %d%% [%f]", ind.r1.Value, ind.r2.Value, ind.value, ind.tolerance, ind.note)
	case eModeParallel3:
		return fmt.Sprintf("%d Ohm // %d Ohm // %d Ohm => %d Ohm %d%% [%f]", ind.r1.Value, ind.r2.Value, ind.r3.Value, ind.value, ind.tolerance, ind.note)
	case eModeParallelAndSerie:
		return fmt.Sprintf("%d Ohm // %d Ohm -- %d Ohm => %d Ohm %d%% [%f]", ind.r1.Value, ind.r2.Value, ind.r3.Value, ind.value, ind.tolerance, ind.note)
	case eModeSeriesOnParallel:
		return fmt.Sprintf("(%d Ohm -- %d Ohm) // %d Ohm => %d Ohm %d%% [%f]", ind.r1.Value, ind.r2.Value, ind.r3.Value, ind.value, ind.tolerance, ind.note)

	}
	return ""
}

func generateIndividual(resistors []Resistor) Individual {
	var ind Individual

	ind.r1 = resistors[rand.Intn(len(resistors))]
	ind.r2 = resistors[rand.Intn(len(resistors))]
	ind.r3 = resistors[rand.Intn(len(resistors))]
	ind.mode = Mode(rand.Intn(int(eModeMax)))

	return ind
}

func evaluate(ind *Individual, target uint32) (uint32, uint32, float64) {
	var value float64
	var tolerance float64

	switch ind.mode {
	case eModeSerial:
		value = (float64)(ind.r1.Value + ind.r2.Value)
		tolerance = math.Sqrt(float64(ind.r1.Tolerance)*float64(ind.r1.Tolerance) + float64(ind.r2.Tolerance)*float64(ind.r2.Tolerance))
	case eModeSerial3:
		value = (float64)(ind.r1.Value + ind.r2.Value + ind.r3.Value)
		tolerance = math.Sqrt(float64(ind.r1.Tolerance)*float64(ind.r1.Tolerance) + float64(ind.r2.Tolerance)*float64(ind.r2.Tolerance) + float64(ind.r3.Tolerance)*float64(ind.r3.Tolerance))
	case eModeParallel:
		value = (1.0 / ((1.0 / (float64)(ind.r1.Value)) + (1.0 / (float64)(ind.r2.Value))))
		tolerance = math.Sqrt(float64(ind.r1.Tolerance)*float64(ind.r1.Tolerance) + float64(ind.r2.Tolerance)*float64(ind.r2.Tolerance))
	case eModeParallel3:
		value = (1.0 / ((1.0 / (float64)(ind.r1.Value)) + (1.0 / (float64)(ind.r2.Value)) + (1.0 / (float64)(ind.r3.Value))))
		tolerance = math.Sqrt(float64(ind.r1.Tolerance)*float64(ind.r1.Tolerance) + float64(ind.r2.Tolerance)*float64(ind.r2.Tolerance) + float64(ind.r3.Tolerance)*float64(ind.r3.Tolerance))
	case eModeParallelAndSerie:
		value = (1.0/((1.0/(float64)(ind.r1.Value))+(1.0/(float64)(ind.r2.Value))) + float64(ind.r3.Value))
		tolerance = math.Sqrt(float64(ind.r1.Tolerance)*float64(ind.r1.Tolerance) + float64(ind.r2.Tolerance)*float64(ind.r2.Tolerance) + float64(ind.r3.Tolerance)*float64(ind.r3.Tolerance))
	case eModeSeriesOnParallel:
		value = (1.0 / ((1.0 / (float64)(ind.r1.Value+ind.r2.Value)) + (1.0 / (float64)(ind.r3.Value))))
		tolerance = math.Sqrt(float64(ind.r1.Tolerance)*float64(ind.r1.Tolerance) + float64(ind.r2.Tolerance)*float64(ind.r2.Tolerance) + float64(ind.r3.Tolerance)*float64(ind.r3.Tolerance))
	}

	note := math.Abs((float64)(target)-value) + 100.0 - tolerance

	return (uint32)(value), (uint32)(tolerance), note
}

func mix(ind1 Individual, ind2 Individual, resistors []Resistor) Individual {
	var n Individual
	var used []Resistor
	var usedFiltered []Resistor

	//Build a list with all used values
	used = append(used, ind1.r1)
	used = append(used, ind1.r2)
	if (ind1.mode == eModeSerial3) || (ind1.mode == eModeParallel3) || (ind1.mode == eModeParallelAndSerie) || (ind1.mode == eModeSeriesOnParallel) {
		used = append(used, ind1.r3)
	}
	used = append(used, ind2.r1)
	used = append(used, ind2.r2)
	if (ind2.mode == eModeSerial3) || (ind2.mode == eModeParallel3) || (ind2.mode == eModeParallelAndSerie) || (ind2.mode == eModeSeriesOnParallel) {
		used = append(used, ind2.r3)
	}

	//Deduplicate
	for _, r := range used {
		duplicate := false
		for i := 0; i < len(usedFiltered); i++ {
			if (r.Value == usedFiltered[i].Value) && (r.Tolerance == usedFiltered[i].Tolerance) {
				duplicate = true
				break
			}
		}

		if !duplicate {
			usedFiltered = append(usedFiltered, r)
		}
	}

	//Yes we may re-use the same value, why not?
	n.r1 = usedFiltered[rand.Intn(len(usedFiltered))]
	n.r2 = usedFiltered[rand.Intn(len(usedFiltered))]
	n.r3 = usedFiltered[rand.Intn(len(usedFiltered))]

	if (ind1.mode == eModeSerial) && (ind2.mode == eModeSerial) {
		n.mode = eModeSerial
		return n
	}
	if (ind1.mode == eModeSerial3) && (ind2.mode == eModeSerial3) {
		n.mode = eModeSerial3
		return n
	}
	if ((ind1.mode == eModeSerial) && (ind2.mode == eModeSerial3)) || ((ind1.mode == eModeSerial3) && (ind2.mode == eModeSerial)) {
		n.mode = eModeSerial3
		return n
	}

	if (ind1.mode == eModeParallel) || (ind2.mode == eModeParallel) {
		n.mode = eModeParallel
		return n
	}
	if (ind1.mode == eModeParallel3) || (ind2.mode == eModeParallel3) {
		n.mode = eModeParallel3
		return n
	}

	//All other cases => combo
	n.mode = Mode(rand.Intn(int(eModeMax)))

	return n
}

func cleanup(population []Individual) (uint32, []Individual) {
	var updated []Individual
	var duplicates uint32
	var previous Individual
	duplicates = 0

	for idx, p := range population {

		//The 2 first values can always be reorder, in all modes
		if p.r1.Value > p.r2.Value {
			tmp := p.r1
			p.r1 = p.r2
			p.r2 = tmp
		}

		if (p.mode == eModeSerial3) || (p.mode == eModeParallel3) {
			if p.r2.Value > p.r3.Value {
				tmp := p.r2
				p.r2 = p.r3
				p.r3 = tmp
			}
			if p.r1.Value > p.r2.Value {
				tmp := p.r1
				p.r1 = p.r2
				p.r2 = tmp
			}
		}
		if idx > 0 {
			if (previous.mode == p.mode) && (previous.r1.Value == p.r1.Value) && (previous.r2.Value == p.r2.Value) {
				if (p.mode == eModeSerial) || (p.mode == eModeParallel) {
					duplicates++
				} else if previous.r3.Value == p.r3.Value {
					duplicates++
				}
			}
		}
		previous = p
		updated = append(updated, p)
	}
	return duplicates, updated
}

func mutate(ind Individual, resistors []Resistor) Individual {
	var rpick *Resistor

	rpick = nil

	//Decide which resistor is going to mutate
	if (ind.mode == eModeSerial) || (ind.mode == eModeParallel) {
		if rand.Intn(2) == 0 {
			rpick = &ind.r1
		} else {
			rpick = &ind.r2
		}
	} else {
		r := rand.Intn(3)
		if r == 0 {
			rpick = &ind.r1
		} else if r == 1 {
			rpick = &ind.r2
		} else {
			rpick = &ind.r3
		}
	}

	//Locate mutating resistor on the list
	idx := 0
	for idx = 0; idx < len(resistors); idx++ {
		if resistors[idx].Value == (*rpick).Value {
			break
		}
	}

	if idx == len(resistors) { //Not found??
		return ind
	}
	if len(resistors) == 1 { //Cannot really mutate
		return ind
	}

	//Mutate to the next one
	if idx == 0 { //That was the first one
		*rpick = resistors[idx+1]
	} else if idx == len(resistors)-1 { //That was the last one
		*rpick = resistors[idx-1]
	} else {
		if rand.Intn(2) == 0 {
			*rpick = resistors[idx-1]
		} else {
			*rpick = resistors[idx+1]
		}
	}

	return ind
}

/*
The bruteforce algorithm that kills all the fun.
*/
func dumbBruteforce(target int32, resistors []Resistor) {

	var closestDistance2 float64
	var best2Str string
	var closestDistance3 float64
	var best3Str string

	closestDistance2 = math.MaxFloat64
	closestDistance3 = math.MaxFloat64

	//Combination of 2
	for _, a := range resistors {
		for _, b := range resistors {
			//2 in series
			total := int32(a.Value + b.Value)
			if math.Abs(float64(total-target)) < closestDistance2 {
				closestDistance2 = math.Abs(float64(total - target))
				best2Str = fmt.Sprintf("[ %d ] -- [ %d ] = %d Ohm", a.Value, b.Value, total)
			}

			//2 in parallel
			total = int32(1.0 / ((1.0 / float64(a.Value)) + (1.0 / float64(b.Value))))
			if math.Abs(float64(total-target)) < closestDistance2 {
				closestDistance2 = math.Abs(float64(total - target))
				best2Str = fmt.Sprintf("[ %d ] // [ %d ] = %d Ohm", a.Value, b.Value, total)
			}
		}
	}

	//Combinations of 3
	for _, a := range resistors {
		for _, b := range resistors {
			for _, c := range resistors {

				//3 in series
				total := int32(a.Value + b.Value + c.Value)
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("[ %d ] -- [ %d ] -- [ %d ] = %d Ohm", a.Value, b.Value, c.Value, total)
				}

				//3 in parallel
				total = int32(1.0 / ((1.0 / float64(a.Value)) + (1.0 / float64(b.Value)) + (1.0 / float64(c.Value))))
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("[ %d ] // [ %d ] // [ %d ] = %d Ohm", a.Value, b.Value, c.Value, total)
				}

				//2 in parallel + 1 series
				total = int32(1.0/((1.0/float64(a.Value))+(1.0/float64(b.Value)))) + int32(c.Value)
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("[ %d ] // [ %d ] -- [ %d ] = %d Ohm", a.Value, b.Value, c.Value, total)
				}

				//2 in series + 1 parallel
				total = int32(1.0 / ((1.0 / float64(a.Value+b.Value)) + (1.0 / float64(c.Value))))
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("([ %d ] -- [ %d ]) // [ %d ] = %d Ohm", a.Value, b.Value, c.Value, total)
				}
			}
		}
	}
	fmt.Println("Best result with 2 resistors:", best2Str)
	fmt.Println("Best result with 3 resistors:", best3Str)
}
