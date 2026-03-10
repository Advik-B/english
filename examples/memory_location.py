# Transpiled from English language source
# Memory Location Example
# Demonstrates the memory location feature
print("=== Memory Location Example ===")
# Declare some variables
x = 42
name = "Alice"
numbers = [1, 2, 3]
# Print their memory locations
print("Variable x value:", x)
print("Variable x location:", hex(id(x)))
print("Variable name value:", name)
print("Variable name location:", hex(id(name)))
print("Variable numbers value:", numbers)
print("Variable numbers location:", hex(id(numbers)))
# Show that different variables have different locations
y = 42
print("Variable y (same value as x):", y)
print("Variable y location:", hex(id(y)))
print("Note: x and y have the same value but different memory locations")
print("=== Done! ===")
