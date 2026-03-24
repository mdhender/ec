# Epimethean Challenge

Warning - this isn't Empyrean Challenge.
It's a reconstruction of the game from the 1978 rulebook.

## Copyright

Empyrean Challenge is owned by James Columbo.
No part of this book may be reproduced in any form without his written permission.

## Empires
Each player controls a single empire.

In order to win the game, your empire must control at least 100 planets, with no other empire controlling more than 50% of that number, for four consecutive turns.

You'll use the web site to upload order files and download turn reports.

You can view your empire's reports at https://epimethean.dev/reports/{empireId}/{year}/{quarter}.

## Cluster
The game takes place in a "star cluster" containing 100 stars.
The cluster is actually a cube with a side length of 31.
Coordinates range from (0,0,0) to (30,30,30).

Every star has between 1 and 10 "planets" orbiting it.

Each planet is assigned to an orbit, which ranges from 1 to 10.
In reports, planets are given a type of **terrestrial**, **gas giant**, or **asteroid belt**.
The type is descriptive - it doesn't have any impact on game play.

A **system** is one or more stars that have the same coordinates.
Since the stars share coordinates, they are displayed in reports with a sequence letter to distinguish them.

### Identifiers
Every object in the game that can be used in an order has a unique identifier (**ID**) assigned to it.
The ID's do not overlap - you won't have the same ID for a star and a ship in a game.

## Order Files
Order files must be valid UTF-8 and must not contain backslashes.
Additionally,
1. Players may include comments in their orders. Comments start with "//" and continue to the end of the line.
2. Fields that contain spaces or special characters must be quoted. For example, "Bob's Dragon" or "Slash //" would both be a single field.

All orders are "scrubbed" before processing.
1. Word's "smart quotes" are converted to straight quotation marks.
2. Invalid UTF-8 characters, backslashes, and HTML/shell meta-characters are converted to spaces.
3. Comments are removed, then leading and trailing spaces are removed.
4. Runs of spaces and tabs are compressed into a single space, even in quoted fields.
5. All characters that are not in a quoted string are converted to lower-case.

## Turns
Turns are simple:
1. Players upload order files.
   1. Your files are scrubbed and checked for "fit and format."
   2. Any obvious errors are reported back immediately. (The file is still submitted; this just gives you a chance to fix errors before the turn processes.) 
2. Orders are parsed and executed.
3. Reports are generated.

## Turn phases

Each turn is processed in phases, in the order shown below.

1. Mining and farming production is calculated.
2. Manufacturing production is calculated.
3. Combat takes place.
4. Set up orders are processed.
5. Dis-assembly orders are processed.
6. Build change orders are entered.
7. Mining change orders are entered.
8. Transfers are processed.
9. Assembly orders are processed.
10. All market and trade station activity takes place.
11. Surveys are carried out.
12. Probe and sensor reports are compiled.
13. Espionage activity takes place.
14. Ship movement occurs.
15. Draft orders are processed.
16. Pay and ration orders are entered.
17. Rebellion occurs.
18. Rebel increases take place.
19. Naming and control orders are processed.
20. Population increases are calculated.
21. News service reports are compiled.

## Orders

Notes:

1. ID is the unique identifier for the object being ordered (this is usually a ship or colony.)
2. Percent Committed must include the percent sign. For example, "18%" is valid; "18" is not. 10% means barely committed and 100% means totally committed.

### Set Up Orders

The Setup order is the only order in the game that spans multiple lines.
You **must** put each transfer on a separate line.

* Set up

    setup {"ship" or "colony"} from {ID}
    transfer {quantity} {unitCode}
    end

The word "end" must be written at the end of setup orders.

Example

    setup ship from 29
        transfer 50,000 ssu
        transfer 5 spac-1
        transfer 5 lsu-1
        transfer 5 food
        transfer 5 pro
        transfer 1 sen-1
        transfer 10,000 fuel
        transfer 61 hype-1
    end

### Combat orders

* Bombard

    {ID} bombard {targetID} {percentCommitted}

* Invade

    {ID} invade {targetID} {percentCommitted}

* Raid

    {ID} raid {targetID} {percentCommitted} {unitCode}

* Support Attacker

    {ID} support {attackerID} {targetID} {percentCommitted}

* Support Defender

    {ID} support {defenderID} {percentCommitted}


### Assembly orders

NOTE: Factories and mines are assembled into groups automatically.
This can cause issues in rare cases.

* Assemble Factory

    {ID} assemble {quantity} {factoryUnitCode} {unitCode}

Example

    91 assemble 54,000 fact-6, cons

* Assemble Mine

    {ID} assemble {quantity} {mineUnitCode} {depositID}

Example

    83 assemble 25,680 mine-2 148

* Assemble Units

    {ID} assemble {quantity} {unitCode}

Example

    58 assemble 6,000 msl-1

## Bugs

    There will be bugs.

Please report bugs on the Discord server.
