# Transpiled from English language source
import math

# Test standard library functions
# Math functions
print("=== Math Functions ===")
result = 0
result = math.sqrt(16)
print("sqrt(16) =", result)
result = math.pow(2, 8)
print("pow(2, 8) =", result)
result = abs(-42)
print("abs(-42) =", result)
# String functions
print("=== String Functions ===")
text = "hello world"
text = text.upper()
print("uppercase:", text)
text = text.lower()
print("lowercase:", text)
# List functions
print("=== List Functions ===")
mylist = [3, 1, 4, 1, 5, 9, 2, 6]
print("Original list:", mylist)
mylist = sorted(mylist)
print("Sorted:", mylist)
mylist = list(reversed(mylist))
print("Reversed:", mylist)
mylist = mylist + [100]
print("After append(100):", mylist)
print("All stdlib functions work!")
