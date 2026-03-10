# Transpiled from English language source
# Test safe imports
print("=== Testing Safe Imports ===")
# Import safely - should NOT print "Initializing library..." or "Library initialized!"
from library_with_init import *

# But functions and variables should be available
result = 0
result = add(5, 3)
print("add(5, 3) =", result)
result = multiply(4, 6)
print("multiply(4, 6) =", result)
print("version =", version)
print("=== Safe import works! ===")
