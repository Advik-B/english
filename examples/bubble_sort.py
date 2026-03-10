# Transpiled from English language source
# Bubble Sort Example
# Implements bubble sort algorithm to sort an array
print("=== Bubble Sort ===")
# Create an unsorted array
arr = [64, 34, 25, 12, 22, 11, 90]
print("Original array:")
for item in arr:
    print(item)
# Bubble sort implementation
n = 7
i = 0
j = 0
current = 0
next_pos = 0
nextval = 0
while i < (n - 1):
    j = 0
    while j < ((n - i) - 1):
        current = arr[j]
        next_pos = j + 1
        nextval = arr[next_pos]
        if current > nextval:
            # Swap elements
            arr[j] = nextval
            arr[next_pos] = current
        j = j + 1
    i = i + 1
print("Sorted array:")
for item in arr:
    print(item)
print("=== Done! ===")
