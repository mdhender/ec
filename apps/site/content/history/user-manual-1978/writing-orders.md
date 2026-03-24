---
title: Writing Orders
weight: 180
---

## General Rules


All orders should be typed or clearly printed on 8.5 by 11 inch paper, and pages should be numbered if there are more than one.

At the top of each page of his turn, the player places his signature, player name, player ID No., game No., and game turn number.

The game turn No. is the number of the last print-out received.

To distinguish between zero and O, the player writes his zero's with a slash through them: 0.

All orders referring to a quantity of a unit do not need a comma between the quantity and the unit.

Orders should be written in the same order as the Sequence of Play; orders not written this way may not be executed.
For instance, if a player writes an assembly order before a dis-assembly order, the computer will decide that there is no assembly order, and the assembly order will not be executed.

All orders are addressed to ships or colonies.
The I.D. No. of the ship or colony is the first part of most orders.

Orders are carried out only if the ship or colony has the necessary units to fulfill the order.
For example, if a player ordered 100 space drives transferred, and he only has 60, then only 60 space dives will be transferred.

Players can use "//" to prefix comments in their orders.
The "//" and all text to then end of the line will be ignored by the parser.
For example,

    // This is a comment
    39, bombard, 121, 75%. // This is a comment

is the same as

    39, bombard, 121, 75%.

## List of Orders


Their format and examples.

### Combat orders


{{< callout type="info" >}}
For all combat orders, the percent commited is expressed as an integer and the percentage sign is required.
10% means barely committed and 100% means totally committed.
{{< /callout >}}

**Bombard Format**

```text
Ship/Colony No. , "bombard" , Defender Ship/Colony No. , percent commited .
```


**Bombard Examples**

```text
39, bombard, 121, 75%.
```


**Invade Format**

```text
Ship/Colony No. , "invade" , Defender Ship/Colony No. , percent commited .
```


**Invade Examples**

```text
22, invade, 342, 55%.
```


**Raid Format**

```text
Ship/Colony No. , "raid" , Defender Ship/Colony No. , percent committed , material raided .
```


{{< callout type="info" >}}
In raids, the material raided must follow.
{{< /callout >}}

**Raid Examples**

```text
98, raid, 644, 28%, gold.
```


{{< callout type="info" >}}
For support orders, the supported players ship/colony ID No. is shown.
{{< /callout >}}

**Support Attacker Format**

```text
Ship/Colony No. , "support" , attacker colony or ship ID No. , Defender Ship/Colony No. , percent commited .
```


**Support Attacker Example**

```text
20, support, 342, 45, 35%.
```


{{< callout type="info" >}}
No. 20 supports No. 342 in its attack on No. 45.
{{< /callout >}}

**Support Defender Format**

```text
Ship/Colony No. , "support" , Defender Ship/Colony No. , percent commited .
```


**Support Defender Examples**

```text
20, support, 342, 40%.
```


{{< callout type="info" >}}
No. 20 supports 342 in its defense.
{{< /callout >}}

### Set Up Orders


**Format**

```text
"set up" , type (ship or colony) , Establishing Ship/Colony No. , "transfer" , quantity and item , ... , "end".

The word "END" must be written at the end of set up orders only.
```


{{< callout type="info" >}}
You may split this order over several lines.
{{< /callout >}}

**Examples**

```text
set up, ship, 29,
  transfer, 50,000 structural units,
                 5 space drives-1,
                 5 life supports-1,
                 5 food,
                 5 professionals,
                 1 sensor-1,
            10,000 fuel,
                61 hyper engines-1,
end.
```


{{< callout type="info" >}}
The "-1" in "space drives-1" refers to TL-1.
{{< /callout >}}

### Assembly orders


{{< callout type="info" >}}
Factories and mines are assembled into groups automatically.
This can cause issues in rare cases.
{{< /callout >}}

**Factory Assembly Format**

```text
Ship/Colony No. , "assemble" , quantity of factory units , units the factory will make .
```


**Factory Assembly Examples**

```text
91, assemble, 54,000 factories-6, consumer goods.
```


**Mine Assembly Format**

```text
Ship/Colony No. , "assemble" , quantity of mine units , Deposit No. .
```


**Mine Assembly Examples**

```text
83, assemble, 25,680 mine-2, 148.
```


**Other Assembly Format**

```text
Ship/Colony No. , "assemble" , quantity of units .
```


**Other Assembly Examples**

```text
58, assemble, 6,000 missile launchers-1.
```


### Disassembly orders


Format and examples are the same as for assembly orders, with the word "disassemble" replacing the word "assemble."

### Build Change Orders


**Format**

```text
Ship/Colony No. , "build change" , Factory Group No. , item to start building (or "retool") .
```


**Examples**

```text
16, build change, 8, energy weapons-4.
16, build change, 9, research.
16, build change, 10, retool.
```


### Transfer Orders


**Format**

```text
Ship/Colony No. , "transfer" , quantity and unit type, Receiving Ship/Colony No.
```


{{< callout type="info" >}}
A future version of this command will allow you to transfer multiple items with one command.
Until then, you must use multiple transfer orders.
{{< /callout >}}

**Examples**

```text
22, transfer, 10 spy units, 29.
```


### Mining Change Orders


**Format**

```text
Ship/Colony No. , "mining" , Mining Group No. , new Deposit No .
```


**Examples**

```text
348, mining, 18, 92.
```


### Market Order


**Buy Units Format**

```text
Ship/Colony No. , "buy" , quantity , unit type, price in GOLD each .
```


**Examples**

```text
555, buy, 25,600, STU, 0.01.
555, buy, 1,300, AUT-1, 1,000.
53, buy, 8,000, space drive-3, 0.1.
```


**Buy Technology Levels Format**

```text
Ship/Colony No. , "buy" , technology level , price in GOLD each .
```


**Examples**

```text
53, buy, TL-6, 1,000,000.
```


{{< callout type="info" >}}
Quantity for TL must be omitted.
{{< /callout >}}

**Sell Units Format**

```text
Ship/Colony No. , "sell" , quantity , unit type , price in GOLD each .
```


**Examples**

```text
721, sell, STU, 0.5.
721, sell, AUT-4, 8,000.
44, sell, 4, space drive-3, 0.2.
```


**Sell Technology Levels Format**

```text
Ship/Colony No. , "sell" , technology level, price in GOLD each .
```


**Examples**

```text
721, sell, TL-4, 800,000.
```


{{< callout type="info" >}}
Quantity for TL must be omitted.
{{< /callout >}}

### Survey Orders


**Format**

```text
Ship/Colony No. , "survey" .
```


**Examples**

```text
23, survey.
```


### Probe Orders


**Format**

```text
Ship/Colony No. , "probe" , Orbit No. , ... .
```


{{< callout type="info" >}}
You may include multiple orbits in one probe order.
They must be separated by commas.
{{< /callout >}}

**Examples**

```text
28, probe, 6.
31, probe, 2, 4, 5.
```


### Spy Orders


**Check Rebel Format**

```text
Ship/Colony No. , quantity of spies , "check rebels" .
```


**Convert Rebels Format**

```text
Ship/Colony No. , quantity of spies , "convert rebels" .
```


**Incite Rebels Format**

```text
Ship/Colony No. , quantity of spies , "incite rebels" , Defender Ship/Colony No. .
```


**Check For Foreign Spies Format**

```text
Ship/Colony No. , quantity of spies , "check for spies" .
```


**Attack Foreign Spies Format**

```text
Ship/Colony No. , quantity of spies , "attack spies" , Defender Ship/Colony No. .
```


**Gather Information Format**

```text
Ship/Colony No. , quantity of spies , "information" , Defender Ship/Colony No. .
```


**Examples**

```text
38, 1, check rebels.
38, 119, convert rebels.
38, 1, check for spies.
38, 102, attack spies, 54.
38, 998, incite rebels, 54.
38, 12, information, 54.
```


### News Release


**Market Planet News Format**

```text
"news" , Market Planet Location , message text .
"news" , Market Planet Location , message text , signature .
```


{{< callout type="info" >}}
The message text must be enclosed in double quotes.
Odd characters in the message may cause your orders to be rejected.
{{< /callout >}}

{{< callout type="info" >}}
Signature is optional. If you include it, it must be enclosed in double quotes.
{{< /callout >}}

**Market Planet News Examples**

```text
news, 02-29-64/3, "I'll be putting space drives-3 on the market next turn.".
news, 02-29-64/3, "I'll be putting space drives-3 on the market next turn.", "Tras-yo of Blenora".
```


**Trade Station News Format**

```text
"news" , Trade Station Colony No. , message text .
"news" , Trade Station Colony No. , message text , signature .
```


{{< callout type="info" >}}
The message text must be enclosed in double quotes.
Odd characters in the message may cause your orders to be rejected.
{{< /callout >}}

{{< callout type="info" >}}
Signature is optional. If you include it, it must be enclosed in double quotes.
{{< /callout >}}

**Trade Station News Examples**

```text
news, 632, "I'll be putting space drives-3 on the market next turn.".
news, 632, "I'll be putting space drives-3 on the market next turn.", "Tras-yo of Blenora".
```


### Jump Orders


**In-System Jump Format**

```text
Ship No. , "move" , Orbit No. .
```


{{< callout type="info" >}}
If your ship is in the 11th orbit you will need to specify which star you are jumping to.
You do that by putting the star's sequence letter and a dash in front of the orbit number.
{{< /callout >}}

**In-System Jump Examples**

```text
77, move, 6.
88, move, C-4.
```


**System Jump Format**

```text
Ship No. , "move" , Destination Location
```


**System Jump Examples**

```text
79, move, 4-6-19.
```


### Draft Orders


**Format**

```text
Ship/Colony No. , "draft" , quantity and type of unit .
```


**Examples**

```text
13, draft, 3,600 soldier.
78, draft, 16,880 trainee.
99, draft, 5,000 CNW.
```


### Disband Orders


**Format**

```text
Ship/Colony No. , "disband" , quantity and type of unit .
```


**Examples**

```text
13, disband, 3,600 soldier.
78, disband, 16,880 trainee.
99, disband, 5,000 CNW.
```


### Pay Orders


**Format**

```text
Ship/Colony No. , "pay" , wages , type .
```


**Examples**

```text
38, pay, 0.7, unskilled.
38, pay, 1.6, professional.
38, pay, 1.2, soldier.
```


### Ration Orders


**Format**

```text
Ship/Colony No. , "ration" ,  ration percentage % .
```


{{< callout type="info" >}}
Ration is expressed as an integer and the percentage sign is required.
50% means one-half of a full ration and 100% means a full ration.
Starvation sets in at 25% of a full ration.
{{< /callout >}}

{{< callout type="info" >}}
All population units on the ship/colony are assigned the same ration.
{{< /callout >}}

**Examples**

```text
16, ration, 50%.
```


### Control Orders


**Format**

```text
Empire No. , "control" , Location .
```


{{< callout type="info" >}}
The location must include the orbit number.
{{< /callout >}}

**Examples**

```text
28, control, 2-4-6/9.
```


### Un-Control Orders


Same as Control orders, except the word "un-control" replaces "control."

**Examples**

```text
28, un-control, 2-4-6/9.
```


### Naming Orders


{{< callout type="info" >}}
Names may be no more than 24 characters long, including blanks.
{{< /callout >}}

#### Naming Planets


**Format**

```text
Location , "name" , name .
```


{{< callout type="info" >}}
The name must be enclosed in double quotes.
Odd characters in the name may cause your orders to be rejected.
{{< /callout >}}

**Examples**

```text
5-12-38, name, "Goldball".
5-12-38/2, name, "Goldball Prime".
```


#### Naming Ships and Colonies


**Format**

```text
Ship/Colony No. , "name" , name
```


{{< callout type="info" >}}
The name must be enclosed in double quotes.
Odd characters in the name may cause your orders to be rejected.
{{< /callout >}}

**Examples**

```text
39, name, "Dragonfire".
```


### Trade Station Orders


**Format**

```text
Trade Station Ship/Colony No. , "permission" , Receiving Ship/Colony No. , "granted" .
Trade Station Ship/Colony No. , "permission" , Receiving Ship/Colony No. , "denied" .
```


{{< callout type="info" >}}
The receiving ship/colony is a proxy. By granting or denying permission to it, you are actually granting/denying permission to the Empire that controls that ship or colony.
{{< /callout >}}

**Examples**

```text
138, permission, 200, granted.
162, permission, 100, denied.
```


### Colonization Orders


**Format**

```text
Receiving Ship/Colony No. , "permission to colonize" , Location .
```


{{< callout type="info" >}}
The location must include the orbit number.
{{< /callout >}}

{{< callout type="info" >}}
The receiving ship/colony is a proxy. By giving permission to it, you are actually granting permission to the Empire that controls that ship or colony. You can't revoke permission to colonize after it has been granted.
{{< /callout >}}

**Examples**

```text
129, permission to colonize, 99-12-26/3.
```
