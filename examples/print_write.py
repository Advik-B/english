# Transpiled from English language source
# Print and Write Example
# Demonstrates enhanced print and write statements
print("=== Print and Write Examples ===")
# Multiple arguments with comma
print("Hello", "World")
# Print with variable
name = "Alice"
print("Hello,", name)
# Print with expression
x = 10
y = 20
print("The sum of", x, "and", y, "is", x + y)
# Write statement (no newline)
print("Loading", end="")
print(".", end="")
print(".", end="")
print(".", end="")
print("\n", end="")
# Write for inline output
items = ["Apple", "Banana", "Cherry"]
print("Items: ", end="")
for item in items:
    print(item, end="")
    print(" ", end="")
print("\n", end="")
# Escape sequences in strings
print("Line 1\nLine 2\nLine 3")
print("Tab\tSeparated\tValues")
print("=== Done! ===")
