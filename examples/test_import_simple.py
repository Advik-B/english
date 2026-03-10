# Transpiled from English language source
# Import with different syntax styles
print("=== Testing different import syntax ===")
# Import using simple syntax
from string_utils import *

# Test imported functions
greeting = ""
greeting = makeGreeting("Alice")
print(greeting)
farewell = ""
farewell = makeFarewell("Bob")
print(farewell)
print("=== Done! ===")
