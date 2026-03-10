# Transpiled from English language source
# Turing Machine Simulation
# This program demonstrates Turing completeness by implementing a simple Turing machine
# that performs binary increment (adds 1 to a binary number)
print("=== Turing Machine Simulation ===")
print("This Turing machine increments a binary number")
# Initialize the tape with a binary number (1011 = 11 in decimal)
# We use 0 and 1, with 2 representing blank
tape = [2, 1, 0, 1, 1, 2, 2, 2, 2, 2]
head = 4
state = 1
steps = 0
# States: 1=find_right, 2=increment, 3=done
print("Initial tape position 4:")
print(tape[4])
# Declare current_symbol before the loop
current_symbol = 0
# Main Turing machine loop
while True:
    current_symbol = tape[head]
    steps = steps + 1
    # State 1: find_right - move to the rightmost digit
    if state == 1:
        if current_symbol == 2:
            head = head - 1
            state = 2
        else:
            head = head + 1
        # State 2: increment - add 1 with carry
    elif state == 2:
        if current_symbol == 0:
            tape[head] = 1
            state = 3
        elif current_symbol == 1:
            tape[head] = 0
            head = head - 1
        else:
            tape[head] = 1
            state = 3
        # State 3: done - halt
    elif state == 3:
        break
    # Safety limit to prevent infinite execution
    if steps > 50:
        break
print("Final result at positions 1-4:")
print(tape[1])
print(tape[2])
print(tape[3])
print(tape[4])
print("Steps taken:")
print(steps)
print("Binary 1011 (11) incremented to 1100 (12)")
print("=== Turing Complete! ===")
