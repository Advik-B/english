# Transpiled from English language source
from typing import Final

# Let Syntax Example
# Demonstrates the various ways to declare variables using 'let'
print("=== Let Syntax Examples ===")
# Basic let syntax
x = 10
print("let x be 10:", x)
# Assignment style
y = 20
print("let y = 20:", y)
# Equal keyword
z = 30
print("let z equal 30:", z)
# Constant with always before be
PI: Final = 3.14159
print("let PI always be 3.14159:", PI)
# Constant with always after be
E: Final = 2.71828
print("let E be always 2.71828:", E)
# Equal to syntax (natural English)
max = 100
print("let max be equal to 100:", max)
# Using let with expressions
sum = (x + y) + z
print("let sum = x + y + z:", sum)
# Scoped variables in loops
print("Scoped variables in loops:")
counter = 0
while counter < 3:
    # This variable is scoped to each loop iteration
    temp = counter * 10
    print("temp:", temp)
    counter = counter + 1
print("=== Done! ===")
