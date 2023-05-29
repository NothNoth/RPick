package main

import (
	_ "crypto/sha512"
	"fmt"
	"math"
	"strconv"
	"syscall/js"
)

func main() {
	done := make(chan struct{}, 0)
	js.Global().Set("rpickbf", js.FuncOf(rpickbf))
	<-done
}

func rpickbf(this js.Value, args []js.Value) interface{} {
	var resistors []int

	if len(args) != 2 {
		return "Invalid call"
	}

	if args[0].Type() != js.TypeString {
		return "Invalid call (target is not a string)"
	}
	if args[1].Type() != js.TypeObject {
		return "Invalid call (resistors list is not an object)"
	}

	target, err := strconv.Atoi(args[0].String())
	if err != nil {
		return "Target value is not a number! " + args[0].Type().String()
	}

	count := args[1].Length()
	for i := 0; i < count; i++ {
		value := args[1].Index(i).String()
		valueI, err := strconv.Atoi(value)
		if err != nil {
			return "Invalid resistor value " + value
		}

		resistors = append(resistors, valueI)
	}

	return dumbBruteforce(int32(target), resistors)
}

func dumbBruteforce(target int32, resistors []int) string {

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
			total := int32(a + b)
			if math.Abs(float64(total-target)) < closestDistance2 {
				closestDistance2 = math.Abs(float64(total - target))
				best2Str = fmt.Sprintf("[ %d ] -- [ %d ] = %d Ohm", a, b, total)
			}

			//2 in parallel
			total = int32(1.0 / ((1.0 / float64(a)) + (1.0 / float64(b))))
			if math.Abs(float64(total-target)) < closestDistance2 {
				closestDistance2 = math.Abs(float64(total - target))
				best2Str = fmt.Sprintf("[ %d ] // [ %d ] = %d Ohm", a, b, total)
			}
		}
	}

	//Combinations of 3
	for _, a := range resistors {
		for _, b := range resistors {
			for _, c := range resistors {

				//3 in series
				total := int32(a + b + c)
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("[ %d ] -- [ %d ] -- [ %d ] = %d Ohm", a, b, c, total)
				}

				//3 in parallel
				total = int32(1.0 / ((1.0 / float64(a)) + (1.0 / float64(b)) + (1.0 / float64(c))))
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("[ %d ] // [ %d ] // [ %d ] = %d Ohm", a, b, c, total)
				}

				//2 in parallel + 1 series
				total = int32(1.0/((1.0/float64(a))+(1.0/float64(b)))) + int32(c)
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("[ %d ] // [ %d ] -- [ %d ] = %d Ohm", a, b, c, total)
				}

				//2 in series + 1 parallel
				total = int32(1.0 / ((1.0 / float64(a+b)) + (1.0 / float64(c))))
				if math.Abs(float64(total-target)) < closestDistance3 {
					closestDistance3 = math.Abs(float64(total - target))
					best3Str = fmt.Sprintf("([ %d ] -- [ %d ]) // [ %d ] = %d Ohm", a, b, c, total)
				}
			}
		}
	}

	var result string
	result += "Best result with 2 resistors:" + best2Str
	result += "<br>"
	result += "Best result with 3 resistors:" + best3Str

	return result
}
