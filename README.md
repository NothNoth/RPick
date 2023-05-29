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


## Bruteforce

At some point it appeared to me that the genetic algorithm was funny, but overkill: we end up trying all combinations and find the best option.
So I implemented a very simple and brutal bruteforce which... just works fine (see "dumbBruteforce" in the code).


## Webassembly

I wanted to migrate the bruteforce code into javascipt to allow a simple webpage to be used to do all that nasty stuff.
It ended up with webassembly from go, to a Twitter Bootstrap page.
