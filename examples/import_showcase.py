# Transpiled from English language source
import math

# Import Syntax Showcase
# Demonstrates all supported import syntax variations in the English language
print("=== Import Syntax Variations ===")
print("")
# Variation 1: Simple import (import everything)
print("1. Simple import syntax (imports everything):")
print("   Import \"library.abc\".")
from math_library import *

result1 = 0
result1 = square(3)
print("   -> square(3) = ", result1)
print("")
# Variation 2: Import with "from"
print("2. Import with 'from' keyword:")
print("   Import from \"library.abc\".")
from string_utils import *

result2 = ""
result2 = makeGreeting("Developer")
print("   -> ", result2)
print("")
# Variation 3: Selective import
print("3. Selective import (import specific items):")
print("   Import square and cube from \"library.abc\".")
print("   (Already imported, using existing functions)")
result3 = 0
result3 = cube(2)
print("   -> cube(2) = ", result3)
print("")
# Variation 4: Import everything explicitly
print("4. Import everything explicitly:")
print("   Import everything from \"library.abc\".")
print("   or: Import all from \"library.abc\".")
print("   (Already imported all from math_library)")
print("")
# Show that all imported content is available
print("=== All imports work together ===")
print("Constants from math_library: pi =", math.pi)
farewell = ""
farewell = makeFarewell("User")
print(farewell)
print("")
print("=== Import feature complete! ===")
