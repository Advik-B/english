# Transpiled from English language source
# Array Operations Example
# Demonstrates array creation, access, and modification
print("=== Array Operations ===")
# Create an array
numbers = [10, 20, 30, 40, 50]
print("Original array: [10, 20, 30, 40, 50]")
# Access elements
print("First element:", numbers[0])
print("Third element:", numbers[2])
print("Last element:", numbers[4])
# Modify an element
numbers[2] = 99
print("After setting position 2 to 99:", numbers[2])
# Iterate through the array
print("All elements:")
for item in numbers:
    print(item)
# Calculate sum of all elements
sum = 0
idx = 0
current = 0
while idx < 5:
    current = numbers[int(idx)]
    sum = sum + current
    idx = idx + 1
print("Sum of all elements:", sum)
# String array example
names = ["Alice", "Bob", "Charlie"]
print("Names in the array:")
for item in names:
    print(item)
print("=== Done! ===")
