# Transpiled from English language source
# Advanced Import Example
# Demonstrates multiple imports and complex interactions
print("=== Advanced Import Example ===")
# Import multiple libraries
from math_library import *

from string_utils import *

# Use functions from both libraries
squared = 0
squared = square(4)
print("4 squared is:", squared)
greeting = ""
greeting = makeGreeting("World")
print(greeting)
# Combine results from different libraries
cubed = 0
cubed = cube(2)
msg = ""
msg = makeFarewell("Numbers")
print(msg)
print("2 cubed is:", cubed)
print("=== All libraries working together! ===")
