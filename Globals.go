package main

const initial_stocks = `Cyclops Industries,31966,1
Champion Intelligence,18803,1
Voyage Technologies,6163,0
Dwarf,52906,1
Phenomenon Enterprises,46994,1
White Wolf Sports,90973,1
Surge Aviation,12540,0
Turtle Co.,27772,0
Greatechnolgies,15117,1
Twisterecords,50189,1
Prodintelligence,84983,0
Solsticetems,30138,0
Freacrosystems,62412,0
Rootechnologies,23221,1
Luckytronics,80268,1
Aces,34272,1
Nymph cast,25578,1
Herb aid,70147,1
Mountain stones,96694,1
Vortex ex,96270,1
Ghost media,472,1
Riddle fly,1314,0
Globe mobile,1474,1
Tulip bit,1073,1
Sail air,811,1`

//Global values
const startingCents = 1000000
const numDays = 5
const totalTicks = numDays * 24 * 60

var gamelist = []*gamestate{}
