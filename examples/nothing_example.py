# Transpiled from English language source
# Nothing (Null) Value Example
# Demonstrates the 'nothing' literal for representing absent values
print("=== Nothing (Null) Value ===")
# Declare a variable with no value
result = None
if result == None:
    print("result has no value (is nothing)")
else:
    print("result has a value")
# Assigning a real value
result = 42
if result == None:
    print("result is still nothing")
else:
    print("result now has a value: ")
    print(result)
# Using nothing as a default/sentinel value
found = None
haystack = [10, 20, 30, 40, 50]
for item in haystack:
    if item == 30:
        found = item
if found != None:
    print("Found the value:")
    print(found)
else:
    print("Value not found.")
print("=== Done! ===")
