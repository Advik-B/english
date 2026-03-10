# Transpiled from English language source
# Demo: Bytecode Caching
# This file demonstrates the __engcache__ feature
print("First import (will create cache):")
from math_library import *

result = 0
result = square(5)
print("Square of 5 =", result)
print("Done! Check the __engcache__ directory for cached bytecode.")
