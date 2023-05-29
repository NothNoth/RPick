# RPick - A genetic algorithm to help you pick to resistors

## Introduction

RPick uses a genetic algorithm to help you choose two resistors in your collection (defined in config.json) in order to obtain a target value.

The algorithm will test combinations among you collection but putting two values in series or parallel and try to match your target.

The tolerance of the combined resistors is computed and integrated into the global score in order to statistically have the best chance to obtain the target value in practice.


## Usage

Edit "config.json" and set your resistors collection (what do you have in stock).

Build:

    go build

Run:

    ./rpick <config file> <target value in Ohm>

Example:

    ./rpick config.json 54321
    [...]
    Generation 100  |  51000Ohm -- 3300Ohm => 54300 Ohm [21.000000]


The best current result is displayed (interrupt with CTRL+C).

  '--' means: resistors in series
  '//' means: resistors in parallel

## Upcoming work

  - Combine 3 resistors
