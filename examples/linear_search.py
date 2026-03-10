# Transpiled from English language source
# Linear Search Example
# Searches for elements in an array
print("=== Linear Search ===")
# Create array to search
arr = [10, 25, 30, 45, 50, 65, 70, 85, 90]
print("Array: 10, 25, 30, 45, 50, 65, 70, 85, 90")
# Function to perform linear search
def linearSearch(arr_size, target):
    idx = 0
    current = 0
    while idx < arr_size:
        current = arr[int(idx)]
        if current == target:
            return idx
        idx = idx + 1
    return -1

# Search for various values
size = 9
print("Searching for 45:")
pos1 = 0
pos1 = linearSearch(size, 45)
if pos1 == -1:
    print("Not found")
else:
    print("Found at index:", pos1)
print("Searching for 100:")
pos2 = 0
pos2 = linearSearch(size, 100)
if pos2 == -1:
    print("Not found")
else:
    print("Found at index:", pos2)
print("Searching for 10:")
pos3 = 0
pos3 = linearSearch(size, 10)
if pos3 == -1:
    print("Not found")
else:
    print("Found at index:", pos3)
print("=== Done! ===")
