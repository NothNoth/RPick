# RPick - A genetic algorithm to help you pick to resistors

## Introduction

RPick uses a genetic algorithm to help you choose two resistors in your collection (defined in config.json) in order to obtain a target value.

The algorithm will test combinations among you collection but putting two values in series or parallel and try to match your target.

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

## Upcoming work

  - Use tolerance values to try to reduce the obtained tolerance
  - Combine 3 resistors
