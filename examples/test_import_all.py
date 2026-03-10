# Transpiled from English language source
# Test import all (synonym for everything)
print("=== Testing Import All ===")
# Import all from the library
from string_utils import *

msg = ""
msg = makeGreeting("World")
print(msg)
msg = makeFarewell("Friend")
print(msg)
print("=== Import all works! ===")
