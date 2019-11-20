package main

const initial_stocks = `Cyclops Industries,31966,1
Champion Intelligence,1880,1
Voyage Technologies,616,0
Dwarf,5290,1
Phenomenon Enterprises,4699,1
White Wolf Sports,9097,1
Surge Aviation,1254,0
Turtle Co.,2777,0
Greatechnolgies,1511,1
Twisterecords,5018,1
Prodintelligence,8498,0
Solsticetems,3013,0
Freacrosystems,6241,0
Rootechnologies,2322,1
Luckytronics,8026,1
Aces,3427,1
Nymph cast,2557,1
Herb aid,701,1
Mountain stones,966,1
Vortex ex,9627,1
Ghost media,472,1
Riddle fly,131,0
Globe mobile,1474,1
Tulip bit,1073,1
Sail air,811,1`

//Global values
const startingCents = 10000000
const numDays = 5
const totalTicks = numDays * 24 * 60

var gamelist = []*gamestate{}
