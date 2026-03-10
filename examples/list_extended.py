# Transpiled from English language source
def _flatten(lst):
    return [item for sublist in lst for item in sublist]

def _unique(lst):
    seen = []
    for item in lst:
        if item not in seen:
            seen.append(item)
    return seen

# List Functions Example
# Demonstrates the extended list standard library
print("=== Extended List Functions ===")
numbers = [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]
print("Original list:")
print(numbers)
print("")
# Sum
print("Sum of all numbers:")
print(sum(numbers))
# Count
print("Count of elements:")
print(len(numbers))
# First and Last
print("First element:")
print(numbers[0])
print("Last element:")
print(numbers[-1])
# Unique (remove duplicates)
print("Unique values:")
print(_unique(numbers))
# Slice
print("Slice from index 2 to 6:")
print(numbers[2:6])
# Flatten nested lists
print("")
print("=== Flatten ===")
matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
print("Original nested list:")
print(matrix)
print("Flattened:")
print(_flatten(matrix))
# Sort
print("")
print("=== Sort and Reverse ===")
unsorted = [64, 25, 12, 22, 11]
print("Unsorted:")
print(unsorted)
unsorted = sorted(unsorted)
print("Sorted:")
print(unsorted)
unsorted = list(reversed(unsorted))
print("Reversed:")
print(unsorted)
# Computing average using sum and count
print("")
print("=== Average ===")
data = [10, 20, 30, 40, 50]
total = sum(data)
n = len(data)
avg = total / n
print("Average of [10, 20, 30, 40, 50]:")
print(avg)
print("=== Done! ===")
