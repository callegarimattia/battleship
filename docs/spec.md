# Battleship Game Specification

## 1. Game Overview

**Objective:** Be the first player to sink all of the opponent's ships.
**Players:** 2 (Human vs. Human or Human vs. AI).
**State:** Turn-based strategy with hidden information (fog of war).

---

## 2. The Game Board

* **Dimensions:** 10 x 10 Grid.
* **Coordinates:**
* **Rows:** Identified by letters **A through J**.
* **Columns:** Identified by numbers **1 through 10**.


* **View:** Each player possesses two grids:
1. **Primary Grid:** Displays their own ships and opponent's attacks.
2. **Tracking Grid:** Blank initially; records the player's shots (hits/misses) against the opponent.



---

## 3. The Fleet (Assets)

Each player receives the same set of 5 ships. They must be placed on the Primary Grid.

| Ship Type | Length (Cells) | Quantity |
| --- | --- | --- |
| **Carrier** | 5 | 1 |
| **Battleship** | 4 | 1 |
| **Cruiser** | 3 | 1 |
| **Submarine** | 3 | 1 |
| **Destroyer** | 2 | 1 |

---

## 4. Logic & Rules

### **Setup Phase**

* **Placement:** Ships can be placed **Horizontally** or **Vertically**.
* **Constraints:**
* **No Overlap:** Ships cannot occupy the same cell.
* **In Bounds:** Ships must fit entirely within the 10x10 grid.
* **No Diagonals:** Diagonal placement is not allowed.


* **Visibility:** Ship positions are hidden from the opponent.

### **Gameplay Phase**

The game is played in alternating turns.

1. **Targeting:** The active player announces a coordinate (e.g., "D-4").
2. **Resolution:**
* **MISS:** The coordinate is empty. Mark as white/empty on the Tracking Grid.
* **HIT:** The coordinate is occupied by a ship. Mark as red/hit on the Tracking Grid.


3. **Sinking:** When every coordinate occupied by a specific ship has been hit, that ship is **SUNK**. The opponent must announce, "You sunk my [Ship Name]!"

### **End Game**

* **Win Condition:** A player wins immediately when **all 17 grid cells** occupied by the opponent's 5 ships have been hit.

---

## 5. Data States (For Development)

For your data structure, each cell on the grid should likely track one of the following states:

1. **Empty (Water)** - Initial state.
2. **Occupied** - Contains a ship segment (hidden from opponent).
3. **Miss** - Shot fired at Empty cell.
4. **Hit** - Shot fired at Occupied cell.

